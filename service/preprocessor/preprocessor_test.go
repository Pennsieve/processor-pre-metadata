package preprocessor

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/pennsieve/processor-pre-metadata/client/paths"
	"github.com/pennsieve/processor-pre-metadata/service/pennsieve"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestRun(t *testing.T) {
	datasetId := uuid.NewString()

	integrationID := uuid.NewString()
	inputDir := t.TempDir()
	outputDir := t.TempDir()
	sessionToken := uuid.NewString()
	expectedFiles := NewExpectedFiles(datasetId).WithModels(
		"7931cbe6-7494-4c0b-95f0-9f4b34edc73b",
		"83964537-46d2-4fb5-9408-0b6262a42a56",
		"bb04a8ce-03c9-4801-a0d9-e35cea53ac1b",
	).WithSchemaRelationships(
		"30e7861f-ebae-4cf8-b9bc-2d6b1ae6008d",
		"2514a023-17fe-4743-af5f-094ed3dd339c",
	).WithSchemaLinkedProperties(
		"bbea65fd-b51f-464a-a5d3-dc228ff408c1",
	).WithProxies(map[string][]string{
		"83964537-46d2-4fb5-9408-0b6262a42a56": {"e79e8d65-b094-4f36-94f2-1553cd84b4a2"},
		"bb04a8ce-03c9-4801-a0d9-e35cea53ac1b": {"a9b9d03b-19b3-4a43-b40e-5673ec955e49", "bcf06e0c-42dc-4ce9-9c70-9ee6865ebc7c"}},
	).WithNoProxies(map[string][]string{
		"7931cbe6-7494-4c0b-95f0-9f4b34edc73b": {"7681b4f8-7d10-4855-8c87-7fef3b408c0b"},
		"bb04a8ce-03c9-4801-a0d9-e35cea53ac1b": {"5b07e038-9829-46c9-b698-bf4efef81341"},
	}).Build(t)
	mockServer := newMockServer(t, integrationID, datasetId, expectedFiles)
	defer mockServer.Close()

	metadataPP := NewMetadataPreProcessor(integrationID, inputDir, outputDir, sessionToken, mockServer.URL, mockServer.URL, defaultRecordsBatchSize)

	require.NoError(t, metadataPP.Run())
	expectedFiles.AssertEqual(t, metadataPP.MetadataPath())

}

type ExpectedFile struct {
	// TestdataPath is the path relative to the testdata directory  (which should be the same as the path relative to the metadata directory in the input directory)
	TestdataPath string
	Bytes        []byte
	Content      any
	// APIPath is the request path the mock server will match against.
	APIPath             string
	QueryParams         url.Values
	ExpectFileNotExists bool
}

func (e ExpectedFile) HandlerFunc(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		require.Equal(t, http.MethodGet, request.Method, "expected method %s for %s, got %s", http.MethodGet, request.URL, request.Method)
		if e.QueryParams != nil {
			require.Equal(t, e.QueryParams, request.URL.Query(), "expected query %s for %s, got %s", e.QueryParams, request.URL, request.URL.Query())
		}
		_, err := writer.Write(e.Bytes)
		require.NoError(t, err)
	}
}

type ExpectedFiles struct {
	DatasetID string
	Files     []ExpectedFile
}

func NewExpectedFiles(datasetID string) *ExpectedFiles {
	return &ExpectedFiles{
		DatasetID: datasetID,
		Files: []ExpectedFile{
			{TestdataPath: paths.SchemaFilePath, APIPath: fmt.Sprintf("/models/v1/datasets/%s/concepts/schema/graph", datasetID)},
			{TestdataPath: paths.RelationshipSchemasFilePath, APIPath: fmt.Sprintf("/models/datasets/%s/relationships", datasetID)},
		},
	}
}

func (e *ExpectedFiles) WithModels(modelIDs ...string) *ExpectedFiles {
	for _, modelID := range modelIDs {
		e.Files = append(e.Files, ExpectedFile{
			TestdataPath: paths.PropertiesFilePath(modelID),
			APIPath:      fmt.Sprintf("/models/v1/datasets/%s/concepts/%s/properties", e.DatasetID, modelID),
		}, ExpectedFile{
			TestdataPath: paths.RecordsFilePath(modelID),
			APIPath:      fmt.Sprintf("/models/v1/datasets/%s/concepts/%s/instances", e.DatasetID, modelID),
			QueryParams:  map[string][]string{"limit": {strconv.Itoa(defaultRecordsBatchSize)}, "offset": {strconv.Itoa(0)}},
		})
	}
	return e
}

func (e *ExpectedFiles) WithSchemaRelationships(schemaRelationshipsIDs ...string) *ExpectedFiles {
	for _, schemaRelationshipID := range schemaRelationshipsIDs {
		e.Files = append(e.Files, ExpectedFile{
			TestdataPath: paths.RelationshipInstancesFilePath(schemaRelationshipID),
			APIPath:      fmt.Sprintf("/models/v1/datasets/%s/relationships/%s/instances", e.DatasetID, schemaRelationshipID),
		})
	}
	return e
}

func (e *ExpectedFiles) WithSchemaLinkedProperties(schemaLinkedPropertyIDs ...string) *ExpectedFiles {
	for _, schemaLinkedPropertyID := range schemaLinkedPropertyIDs {
		e.Files = append(e.Files, ExpectedFile{
			TestdataPath: paths.LinkedPropertyInstancesFilePath(schemaLinkedPropertyID),
			APIPath:      fmt.Sprintf("/models/v1/datasets/%s/relationships/%s/instances", e.DatasetID, schemaLinkedPropertyID),
		})
	}
	return e
}

func (e *ExpectedFiles) WithProxies(modelIDToRecordIDs map[string][]string) *ExpectedFiles {
	for modelID, recordIDs := range modelIDToRecordIDs {
		for _, recordID := range recordIDs {
			e.Files = append(e.Files, ExpectedFile{
				TestdataPath: paths.ProxyInstancesFilePath(modelID, recordID),
				APIPath:      fmt.Sprintf("/models/datasets/%s/concepts/%s/instances/%s/files", e.DatasetID, modelID, recordID),
			})
		}
	}
	return e
}

func (e *ExpectedFiles) WithNoProxies(modelIDToRecordIDs map[string][]string) *ExpectedFiles {
	for modelID, recordIDs := range modelIDToRecordIDs {
		for _, recordID := range recordIDs {
			e.Files = append(e.Files, ExpectedFile{
				TestdataPath:        paths.ProxyInstancesFilePath(modelID, recordID),
				APIPath:             fmt.Sprintf("/models/datasets/%s/concepts/%s/instances/%s/files", e.DatasetID, modelID, recordID),
				Bytes:               json.RawMessage("[]"),
				ExpectFileNotExists: true,
			})
		}
	}
	return e
}

func (e *ExpectedFiles) Build(t *testing.T) *ExpectedFiles {
	for i := range e.Files {
		expected := &e.Files[i]
		if !expected.ExpectFileNotExists {
			file := filepath.Join("testdata", expected.TestdataPath)
			bytes, err := os.ReadFile(file)
			require.NoError(t, err)
			expected.Bytes = bytes
			require.NoError(t, json.Unmarshal(bytes, &expected.Content))
		}
	}
	return e
}

func (e *ExpectedFiles) AssertEqual(t *testing.T, actualDir string) {
	for _, expectedFile := range e.Files {
		actualFilePath := filepath.Join(actualDir, expectedFile.TestdataPath)
		if expectedFile.ExpectFileNotExists {
			assert.NoFileExists(t, actualFilePath)
		} else {
			actualBytes, err := os.ReadFile(actualFilePath)
			if assert.NoError(t, err) {
				// Comparing content, not bytes, since checked in expected files may be JSON formatted, while response
				// from mock server is not, making the []bytes unequal.
				var actualContent any
				require.NoError(t, json.Unmarshal(actualBytes, &actualContent))
				assert.Equal(t, expectedFile.Content, actualContent, "actual content %s does not match expected content %s", actualFilePath, expectedFile.TestdataPath)
			}
		}
	}
}

func newMockServer(t *testing.T, integrationID string, datasetID string, expectedFiles *ExpectedFiles) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc(fmt.Sprintf("/integrations/%s", integrationID), func(writer http.ResponseWriter, request *http.Request) {
		require.Equal(t, http.MethodGet, request.Method, "expected method %s for %s, got %s", http.MethodGet, request.URL, request.Method)
		integration := pennsieve.Integration{
			Uuid:          uuid.NewString(),
			ApplicationID: 0,
			DatasetNodeID: datasetID,
			PackageIDs:    nil,
			Params:        nil,
		}
		integrationResponse, err := json.Marshal(integration)
		require.NoError(t, err)
		_, err = writer.Write(integrationResponse)
		require.NoError(t, err)
	})
	for _, expectedFile := range expectedFiles.Files {
		mux.HandleFunc(expectedFile.APIPath, expectedFile.HandlerFunc(t))
	}
	mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		require.Fail(t, "unexpected call to Pennsieve", "%s %s", request.Method, request.URL)
	})
	return httptest.NewServer(mux)
}
