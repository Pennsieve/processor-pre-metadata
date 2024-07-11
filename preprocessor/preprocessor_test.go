package preprocessor

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/pennsieve/processor-pre-metadata/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"testing"
)

func TestRun(t *testing.T) {
	datasetId := uuid.NewString()

	integrationID := uuid.NewString()
	baseDir := t.TempDir()
	sessionToken := uuid.NewString()
	expectedFiles := NewExpectedFiles(datasetId).WithPropFiles(
		"7931cbe6-7494-4c0b-95f0-9f4b34edc73b",
		"83964537-46d2-4fb5-9408-0b6262a42a56",
		"bb04a8ce-03c9-4801-a0d9-e35cea53ac1b",
	).Build(t)
	mockServer := newMockServer(t, integrationID, datasetId, expectedFiles)
	defer mockServer.Close()

	metadataPP := NewMetadataPreProcessor(integrationID, baseDir, sessionToken, mockServer.URL, mockServer.URL)

	currentUser, err := user.Current()
	require.NoError(t, err)
	uid, err := strconv.Atoi(currentUser.Uid)
	require.NoError(t, err)
	gid, err := strconv.Atoi(currentUser.Gid)
	require.NoError(t, err)

	require.NoError(t, metadataPP.Run(uid, gid))
	assert.DirExists(t, metadataPP.InputDirectory())
	assert.DirExists(t, metadataPP.OutputDirectory())
	expectedFiles.AssertEqual(t, metadataPP.MetadataDirectory())

}

type ExpectedFile struct {
	// TestdataPath is the path relative to the testdata directory
	TestdataPath string
	Bytes        []byte
	// APIPath is the request path the mock server will match against.
	APIPath string
}

func (e ExpectedFile) HandlerFunc(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		require.Equal(t, http.MethodGet, request.Method, "expected method %s for %s, got %s", http.MethodGet, request.URL, request.Method)
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
		Files:     []ExpectedFile{{TestdataPath: schemaFileName, APIPath: fmt.Sprintf("/models/v1/datasets/%s/concepts/schema/graph", datasetID)}},
	}
}

func (e *ExpectedFiles) WithPropFiles(modelIDs ...string) *ExpectedFiles {
	for _, modelID := range modelIDs {
		e.Files = append(e.Files, ExpectedFile{
			TestdataPath: propertiesFileName(modelID),
			APIPath:      fmt.Sprintf("/models/v1/datasets/%s/concepts/%s/properties", e.DatasetID, modelID),
		})
	}
	return e
}

func (e *ExpectedFiles) Build(t *testing.T) *ExpectedFiles {
	for i := range e.Files {
		expected := &e.Files[i]
		file := filepath.Join("testdata", expected.TestdataPath)
		bytes, err := os.ReadFile(file)
		require.NoError(t, err)
		expected.Bytes = bytes
	}
	return e
}

func (e *ExpectedFiles) AssertEqual(t *testing.T, actualDir string) {
	for _, expectedFile := range e.Files {
		base := filepath.Base(expectedFile.TestdataPath)
		actualFilePath := filepath.Join(actualDir, base)
		actualBytes, err := os.ReadFile(actualFilePath)
		if assert.NoError(t, err) {
			assert.Equal(t, expectedFile.Bytes, actualBytes, "actual %s does not match expected %s", actualFilePath, expectedFile.TestdataPath)
		}
	}
}

func newMockServer(t *testing.T, integrationID string, datasetID string, expectedFiles *ExpectedFiles) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc(fmt.Sprintf("/integrations/%s", integrationID), func(writer http.ResponseWriter, request *http.Request) {
		require.Equal(t, http.MethodGet, request.Method, "expected method %s for %s, got %s", http.MethodGet, request.URL, request.Method)
		integration := models.Integration{
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
