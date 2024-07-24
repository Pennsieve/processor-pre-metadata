package preprocessor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pennsieve/processor-pre-metadata/logging"
	"github.com/pennsieve/processor-pre-metadata/models/schema"
	"github.com/pennsieve/processor-pre-metadata/pennsieve"
	"github.com/pennsieve/processor-pre-metadata/util"
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

func (m *MetadataPreProcessor) Run() error {
	// get integration info
	integration, err := m.Pennsieve.GetIntegration(m.IntegrationID)
	if err != nil {
		return err
	}
	datasetID := integration.DatasetNodeID
	logger.Info("got integration", slog.String("datasetID", datasetID))

	if err := m.MkDirectories(); err != nil {
		return err
	}
	metadataPath := m.MetadataPath()
	graphModels, err := m.WriteGraphSchema(metadataPath, datasetID)
	if err != nil {
		return err
	}
	if err := m.WriteRecords(metadataPath, datasetID, graphModels); err != nil {
		return err
	}
	return nil
}

func (m *MetadataPreProcessor) WriteGraphSchema(metadataDirectory string, datasetID string) ([]*schema.Model, error) {

	res, err := m.Pennsieve.GetGraphSchema(datasetID)
	if err != nil {
		return nil, err
	}

	graphSchemaFilePath := filepath.Join(metadataDirectory, schemaFilePath)
	var graphSchema []map[string]any
	if err := WriteAndDecodeResponse(res, graphSchemaFilePath, &graphSchema); err != nil {
		return nil, fmt.Errorf("error writing/decoding graph schema: %w", err)
	} else {
		logger.Info("wrote graph schema",
			slog.String("path", graphSchemaFilePath))
	}
	var graphModels []*schema.Model
	schemaRelationshipsBySource := map[string][]schema.Relationship{}
	linkedPropertiesBySource := map[string][]schema.LinkedProperty{}
	for _, schemaElementAsMap := range graphSchema {
		schemaElement, err := schema.FromMap(schemaElementAsMap)
		if err != nil {
			return nil, err
		}
		switch e := schemaElement.(type) {
		case *schema.Model:
			graphModels = append(graphModels, e)
		case *schema.Relationship:
			schemaRelationshipsBySource[e.From] = append(schemaRelationshipsBySource[e.From], *e)
		case *schema.LinkedProperty:
			linkedPropertiesBySource[e.From] = append(linkedPropertiesBySource[e.From], *e)
		default:
			return nil, fmt.Errorf("unknown schema element type: %T", e)
		}
	}
	for _, model := range graphModels {
		modelLogger := model.Logger(logger)
		model.Relationships = schemaRelationshipsBySource[model.ID]
		model.LinkedProperties = linkedPropertiesBySource[model.ID]
		if propRes, err := m.Pennsieve.GetProperties(datasetID, model.ID); err != nil {
			return nil, fmt.Errorf("error getting model %s properties: %w", model.ID, err)
		} else {
			modelPropFilePath := filepath.Join(metadataDirectory, propertiesFilePath(model.ID))
			if err := WriteAndDecodeResponse(propRes, modelPropFilePath, &model.Properties); err != nil {
				return nil, fmt.Errorf("error writing/decoding model %s properties to %s: %w", model.ID, modelPropFilePath, err)
			} else {
				modelLogger.Info("wrote model properties",
					slog.String("path", modelPropFilePath))
			}
		}

	}

	return graphModels, nil
}

func (m *MetadataPreProcessor) WriteRecords(metadataDirectory string, datasetID string, graphModels []*schema.Model) error {
	for _, model := range graphModels {
		modelLogger := model.Logger(logger)
		if recordRes, err := m.Pennsieve.GetAllRecords(datasetID, model.ID, m.RecordsBatchSize); err != nil {
			return err
		} else {
			recordsFilePath := filepath.Join(metadataDirectory, recordsFileName(model.ID))
			if recordsSz, err := WriteJSON(recordsFilePath, recordRes); err != nil {
				return fmt.Errorf("error writing/decoding model %s records to %s: %w", model.ID, recordsFilePath, err)
			} else {
				modelLogger.Info("wrote model records", slog.String("path", recordsFilePath),
					slog.Int64("size", recordsSz))
			}
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
