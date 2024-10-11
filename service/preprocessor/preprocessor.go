package preprocessor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pennsieve/processor-pre-metadata/client/models/schema"
	"github.com/pennsieve/processor-pre-metadata/client/paths"
	"github.com/pennsieve/processor-pre-metadata/service/logging"
	"github.com/pennsieve/processor-pre-metadata/service/pennsieve"
	"github.com/pennsieve/processor-pre-metadata/service/util"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
)

var logger = logging.PackageLogger("preprocessor")

const defaultRecordsBatchSize = 1000

type MetadataPreProcessor struct {
	IntegrationID    string
	DatasetID        string
	InputDirectory   string
	OutputDirectory  string
	Pennsieve        *pennsieve.Session
	RecordsBatchSize int
}

func NewMetadataPreProcessor(integrationID string,
	inputDirectory string,
	outputDirectory string,
	sessionToken string,
	apiHost string,
	api2Host string,
	recordsBatchSize int) *MetadataPreProcessor {
	recordsBatch := recordsBatchSize
	if recordsBatch == 0 {
		recordsBatch = defaultRecordsBatchSize
	}
	return &MetadataPreProcessor{
		IntegrationID:    integrationID,
		InputDirectory:   inputDirectory,
		OutputDirectory:  outputDirectory,
		Pennsieve:        pennsieve.NewSession(sessionToken, apiHost, api2Host),
		RecordsBatchSize: recordsBatch,
	}
}

func FromEnv() (*MetadataPreProcessor, error) {
	integrationID, err := LookupRequiredEnvVar("INTEGRATION_ID")
	if err != nil {
		return nil, err
	}
	inputDirectory, err := LookupRequiredEnvVar("INPUT_DIR")
	if err != nil {
		return nil, err
	}
	outputDirectory, err := LookupRequiredEnvVar("OUTPUT_DIR")
	if err != nil {
		return nil, err
	}
	sessionToken, err := LookupRequiredEnvVar("SESSION_TOKEN")
	if err != nil {
		return nil, err
	}
	apiHost, err := LookupRequiredEnvVar("PENNSIEVE_API_HOST")
	if err != nil {
		return nil, err
	}
	api2Host, err := LookupRequiredEnvVar("PENNSIEVE_API_HOST2")
	if err != nil {
		return nil, err
	}
	return NewMetadataPreProcessor(integrationID, inputDirectory, outputDirectory, sessionToken, apiHost, api2Host, 0), nil
}

func (m *MetadataPreProcessor) WithDatasetID(datasetID string) *MetadataPreProcessor {
	m.DatasetID = datasetID
	return m
}

func (m *MetadataPreProcessor) Run() error {
	if len(m.DatasetID) == 0 {
		// get integration info
		logger.Info("looking up integration", slog.String("integrationID", m.IntegrationID))

		integration, err := m.Pennsieve.GetIntegration(m.IntegrationID)
		if err != nil {
			return err
		}
		datasetID := integration.DatasetNodeID
		m.DatasetID = datasetID
	}
	logger.Info("getting metadata for dataset", slog.String("datasetID", m.DatasetID))
	if err := m.MkDirectories(); err != nil {
		return err
	}
	metadataPath := m.MetadataPath()
	schemaElements, err := m.WriteGraphSchema(metadataPath, m.DatasetID)
	if err != nil {
		return err
	}
	if err := m.WriteInstances(metadataPath, m.DatasetID, schemaElements); err != nil {
		return err
	}
	return nil
}

func (m *MetadataPreProcessor) WriteGraphSchema(metadataDirectory string, datasetID string) (schema.Elements, error) {
	// These don't need to be returned in the schema elements. Most will also appear
	// in the graph schema below and they will be returned from there. Only one that doesn't is the special package proxy
	// relationship which does not need to be included.
	if err := m.WriteRelationshipSchemas(metadataDirectory, datasetID); err != nil {
		return schema.Elements{}, err
	}
	res, err := m.Pennsieve.GetGraphSchema(datasetID)
	if err != nil {
		return schema.Elements{}, err
	}
	graphSchemaFilePath := filepath.Join(metadataDirectory, paths.SchemaFilePath)
	var graphSchema []map[string]any
	if err := WriteAndDecodeResponse(res, graphSchemaFilePath, &graphSchema); err != nil {
		return schema.Elements{}, fmt.Errorf("error writing/decoding graph schema: %w", err)
	} else {
		logger.Info("wrote graph schema",
			slog.String("path", graphSchemaFilePath))
	}

	schemaElements := schema.Elements{}
	for _, schemaElementAsMap := range graphSchema {
		schemaElement, err := schema.FromMap(schemaElementAsMap)
		if err != nil {
			return schema.Elements{}, err
		}
		switch e := schemaElement.(type) {
		case *schema.Model:
			if err := m.WriteProperties(metadataDirectory, datasetID, e); err != nil {
				return schema.Elements{}, err
			}
			schemaElements.Models = append(schemaElements.Models, *e)
		case *schema.Relationship:
			schemaElements.Relationships = append(schemaElements.Relationships, *e)
		case *schema.LinkedProperty:
			schemaElements.LinkedProperties = append(schemaElements.LinkedProperties, *e)
		default:
			return schema.Elements{}, fmt.Errorf("unknown schema element type: %T", e)
		}
	}
	return schemaElements, nil
}

// WriteRelationshipSchemas is a hack to get the special `belongs_to` package proxy relationship schema which is not included in graphSchemaFilePath.
// The other relationship schemas retrieved will be duplicates of the info in graphSchemaFilePath.
func (m *MetadataPreProcessor) WriteRelationshipSchemas(metadataDirectory string, datasetID string) error {
	res, err := m.Pennsieve.GetRelationshipSchemas(datasetID)
	if err != nil {
		return err
	}
	relationshipSchemaFilePath := filepath.Join(metadataDirectory, paths.RelationshipSchemasFilePath)
	writtenCount, err := WriteResponse(res, relationshipSchemaFilePath)
	if err != nil {
		return fmt.Errorf("error writing/decoding relationship schemas: %w", err)
	}
	logger.Info("wrote relationship schemas",
		slog.String("path", relationshipSchemaFilePath),
		slog.Int64("size", writtenCount))
	return nil
}

func (m *MetadataPreProcessor) WriteProperties(metadataDirectory string, datasetID string, model *schema.Model) error {
	modelLogger := model.Logger(logger)
	if propRes, err := m.Pennsieve.GetProperties(datasetID, model.ID); err != nil {
		return fmt.Errorf("error getting model %s properties: %w", model.ID, err)
	} else {
		modelPropFilePath := filepath.Join(metadataDirectory, paths.PropertiesFilePath(model.ID))
		if err := WriteAndDecodeResponse(propRes, modelPropFilePath, &model.Properties); err != nil {
			return fmt.Errorf("error writing/decoding model %s properties to %s: %w", model.ID, modelPropFilePath, err)
		} else {
			modelLogger.Info("wrote model properties",
				slog.String("path", modelPropFilePath))
		}
	}
	return nil
}

func (m *MetadataPreProcessor) WriteInstances(metadataDirectory string, datasetID string, schemaElements schema.Elements) error {
	// Write the records and any package proxies
	for _, model := range schemaElements.Models {
		modelLogger := model.Logger(logger)
		recordRes, err := m.Pennsieve.GetAllRecords(datasetID, model.ID, m.RecordsBatchSize)
		if err != nil {
			return err
		}
		recordsFilePath := filepath.Join(metadataDirectory, paths.RecordsFilePath(model.ID))
		recordsSz, err := WriteJSON(recordsFilePath, recordRes)
		if err != nil {
			return fmt.Errorf("error writing/decoding model %s records to %s: %w", model.ID, recordsFilePath, err)
		}
		modelLogger.Info("wrote model records", slog.String("path", recordsFilePath),
			slog.Int64("size", recordsSz))
		if err := m.WriteProxies(metadataDirectory, datasetID, model.ID, recordRes); err != nil {
			return err
		}

	}

	// Write the relationship instances
	for _, schemaRelationship := range schemaElements.Relationships {
		relLogger := schemaRelationship.Logger(logger)
		relRes, err := m.Pennsieve.GetRelationshipInstances(datasetID, schemaRelationship.ID)
		if err != nil {
			return err
		}
		relationshipInstanceFilePath := filepath.Join(metadataDirectory, paths.RelationshipInstancesFilePath(schemaRelationship.ID))
		if relSz, err := WriteResponse(relRes, relationshipInstanceFilePath); err != nil {
			return fmt.Errorf("error writing/decoding relationship %s instances to %s: %w", schemaRelationship.ID, relationshipInstanceFilePath, err)
		} else {
			relLogger.Info("wrote relationship instances",
				slog.String("path", relationshipInstanceFilePath),
				slog.Int64("size", relSz))
		}
	}

	// Write the linked property instances
	for _, schemaLinkedProperties := range schemaElements.LinkedProperties {
		linkedPropLogger := schemaLinkedProperties.Logger(logger)
		// Using the RelationshipInstances here because linked props are modeled as relationships server side.
		// There is a special linked prop instance endpoint, but it's done by record instead of by schema linked prop id, so
		// its kind of awkward for the layout we've chosen here.
		linkedPropRes, err := m.Pennsieve.GetRelationshipInstances(datasetID, schemaLinkedProperties.ID)
		if err != nil {
			return err
		}
		linkedPropertyInstanceFilePath := filepath.Join(metadataDirectory, paths.LinkedPropertyInstancesFilePath(schemaLinkedProperties.ID))
		if relSz, err := WriteResponse(linkedPropRes, linkedPropertyInstanceFilePath); err != nil {
			return fmt.Errorf("error writing/decoding linked property %s instances to %s: %w", schemaLinkedProperties.ID, linkedPropertyInstanceFilePath, err)
		} else {
			linkedPropLogger.Info("wrote linked property instances",
				slog.String("path", linkedPropertyInstanceFilePath),
				slog.Int64("size", relSz))
		}
	}
	return nil
}

func (m *MetadataPreProcessor) WriteProxies(metadataDirectory, datasetID, modelID string, records []map[string]any) error {
	for _, record := range records {
		recordID, err := GetID(record)
		if err != nil {
			return err
		}
		recordLogger := logger.With(slog.String("recordID", recordID))
		proxies, err := m.Pennsieve.GetProxyInstancesForRecord(datasetID, modelID, recordID)
		if err != nil {
			return err
		}
		if len(proxies) == 0 {
			recordLogger.Info("no proxy instances for record")
		} else {
			proxyInstanceFilePath := filepath.Join(metadataDirectory, paths.ProxyInstancesFilePath(modelID, recordID))
			directory := filepath.Dir(proxyInstanceFilePath)
			if err := os.MkdirAll(directory, 0755); err != nil {
				return fmt.Errorf("error creating proxy instance directory %s: %w", directory, err)
			}
			sz, err := WriteJSON(proxyInstanceFilePath, proxies)
			if err != nil {
				return fmt.Errorf("error writing/decoding proxy instances for %s to %s: %w", recordID, proxyInstanceFilePath, err)
			}
			recordLogger.Info("wrote proxy instances",
				slog.String("path", proxyInstanceFilePath),
				slog.Int64("count", sz),
			)
		}
	}
	return nil
}

func WriteAndDecodeResponse(response *http.Response, filePath string, v any) error {
	defer util.CloseAndWarn(response)

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("error creating file %s: %w", filePath, err)
	}
	tee := io.TeeReader(response.Body, file)
	decoder := json.NewDecoder(tee)
	if err = decoder.Decode(&v); err != nil {
		return fmt.Errorf("error decoding %s %s response: %w",
			response.Request.Method,
			response.Request.URL,
			err)
	}
	return nil
}

func WriteResponse(response *http.Response, filePath string) (int64, error) {
	defer util.CloseAndWarn(response)

	file, err := os.Create(filePath)
	if err != nil {
		return 0, fmt.Errorf("error creating file %s: %w", filePath, err)
	}
	written, err := io.Copy(file, response.Body)
	if err != nil {
		return 0, fmt.Errorf("error writing %s %s response to %s: %w",
			response.Request.Method,
			response.Request.URL,
			filePath,
			err)
	}
	return written, nil
}

func WriteJSON(filePath string, v any) (int64, error) {
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return 0, fmt.Errorf("error marshalling JSON value %s to bytes: %w", v, err)
	}
	file, err := os.Create(filePath)
	if err != nil {
		return 0, fmt.Errorf("error creating file %s: %w", filePath, err)
	}
	written, err := io.Copy(file, bytes.NewReader(jsonBytes))
	if err != nil {
		return 0, fmt.Errorf("error writing JSON value to file %s: %w", filePath, err)
	}
	return written, nil
}

func LookupRequiredEnvVar(key string) (string, error) {
	value := os.Getenv(key)
	if len(value) == 0 {
		return "", fmt.Errorf("no %s set", key)
	}
	return value, nil
}

func GetID(jsonMap map[string]any) (string, error) {
	idAny, inResponse := jsonMap["id"]
	if !inResponse {
		return "", fmt.Errorf("id not found")
	}
	id, isString := idAny.(string)
	if !isString {
		return "", fmt.Errorf("id %s is not string: %T", id, id)
	}
	return id, nil
}
