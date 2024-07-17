package preprocessor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/pennsieve/processor-pre-metadata/logging"
	"github.com/pennsieve/processor-pre-metadata/models"
	"github.com/pennsieve/processor-pre-metadata/pennsieve"
	"github.com/pennsieve/processor-pre-metadata/util"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
)

var logger = logging.PackageLogger("preprocessor")

const schemaFileName = "graphSchema.json"
const defaultRecordsBatchSize = 1000

type MetadataPreProcessor struct {
	IntegrationID    string
	BaseDirectory    string
	Pennsieve        *pennsieve.Session
	RecordsBatchSize int
}

func NewMetadataPreProcessor(integrationID string,
	baseDirectory string,
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
		BaseDirectory:    baseDirectory,
		Pennsieve:        pennsieve.NewSession(sessionToken, apiHost, api2Host),
		RecordsBatchSize: recordsBatch,
	}
}

func FromEnv() *MetadataPreProcessor {
	integrationID := os.Getenv("INTEGRATION_ID")
	baseDir := os.Getenv("BASE_DIR")
	if integrationID == "" {
		id := uuid.New()
		integrationID = id.String()
	}
	if baseDir == "" {
		baseDir = "/mnt/efs"
	}
	sessionToken := os.Getenv("SESSION_TOKEN")
	apiHost := os.Getenv("PENNSIEVE_API_HOST")
	api2Host := os.Getenv("PENNSIEVE_API_HOST2")
	return NewMetadataPreProcessor(integrationID, baseDir, sessionToken, apiHost, api2Host, 0)
}

func (m *MetadataPreProcessor) Run(uid int, gid int) error {
	// create subdirectories
	// inputDir
	inputDir, err := m.MkInputDirectory()
	if err != nil {
		return err
	}
	logger.Info("created input directory", slog.String("path", inputDir))

	// outputDir
	outputDir, err := m.MkOutputDirectory(uid, gid)
	if err != nil {
		return err
	}
	logger.Info("created output directory", slog.String("path", outputDir))

	// get integration info
	integration, err := m.Pennsieve.GetIntegration(m.IntegrationID)
	if err != nil {
		return err
	}
	datasetID := integration.DatasetNodeID
	if datasetID == "" {
		datasetID = "N:dataset:e323328c-13c3-44f3-aaff-4fd5a941ded5"
	}
	logger.Info("got integration", slog.String("datasetID", datasetID))

	metadataDirectory := m.MetadataDirectory()
	if err := os.MkdirAll(metadataDirectory, 0755); err != nil {
		return fmt.Errorf("error creating directory %s: %w", metadataDirectory, err)
	}
	graphModels, err := m.WriteGraphSchema(metadataDirectory, datasetID)
	if err != nil {
		return err
	}
	if err := m.WriteRecords(metadataDirectory, datasetID, graphModels); err != nil {
		return err
	}
	return nil
}

func (m *MetadataPreProcessor) InputDirectory() string {
	return filepath.Join(m.BaseDirectory, "input", m.IntegrationID)
}

func (m *MetadataPreProcessor) MkInputDirectory() (string, error) {
	inputDir := m.InputDirectory()
	err := os.MkdirAll(inputDir, 0755)
	if err != nil {
		return "", fmt.Errorf("error creating input directory %s: %w", inputDir, err)
	}
	return inputDir, nil
}

func (m *MetadataPreProcessor) BaseOutputDirectory() string {
	return filepath.Join(m.BaseDirectory, "output")
}

func (m *MetadataPreProcessor) OutputDirectory() string {
	return filepath.Join(m.BaseOutputDirectory(), m.IntegrationID)
}

func (m *MetadataPreProcessor) MetadataDirectory() string {
	return filepath.Join(m.InputDirectory(), "metadata")
}

func (m *MetadataPreProcessor) MkOutputDirectory(uid, gid int) (string, error) {
	baseOutputDir := m.BaseOutputDirectory()
	outputDir := m.OutputDirectory()
	err := os.MkdirAll(outputDir, 0777)
	if err != nil {
		return "", fmt.Errorf("error creating output directory %s: %w", outputDir, err)
	}
	err = filepath.WalkDir(baseOutputDir, func(name string, info os.DirEntry, err error) error {
		if err == nil {
			err = os.Chown(name, uid, gid)
		}
		return err
	})
	if err != nil {
		return "", fmt.Errorf("error changing owner of output directory %s: %w", outputDir, err)
	}
	return outputDir, err
}

func propertiesFileName(modelID string) string {
	return fmt.Sprintf("%s-properties.json", modelID)
}
func (m *MetadataPreProcessor) WriteGraphSchema(metadataDirectory string, datasetID string) ([]models.Model, error) {

	res, err := m.Pennsieve.GetGraphSchema(datasetID)
	if err != nil {
		return nil, err
	}

	graphSchemaFilePath := filepath.Join(metadataDirectory, schemaFileName)
	var graphSchema []map[string]any
	if err := WriteAndDecodeResponse(res, graphSchemaFilePath, &graphSchema); err != nil {
		return nil, fmt.Errorf("error writing/decoding graph schema: %w", err)
	} else {
		logger.Info("wrote graph schema",
			slog.String("path", graphSchemaFilePath))
	}
	var graphModels []models.Model
	for _, schemaElement := range graphSchema {
		if model := models.ModelFromMap(schemaElement); model != nil {
			graphModels = append(graphModels, *model)
			modelLogger := model.Logger(logger)
			if propRes, err := m.Pennsieve.GetProperties(datasetID, model.ID); err != nil {
				return nil, fmt.Errorf("error getting model %s properties: %w", model.ID, err)
			} else {
				modelPropFilePath := filepath.Join(metadataDirectory, propertiesFileName(model.ID))
				var props []map[string]any
				if err := WriteAndDecodeResponse(propRes, modelPropFilePath, &props); err != nil {
					return nil, fmt.Errorf("error writing/decoding model %s properties to %s: %w", model.ID, modelPropFilePath, err)
				} else {
					modelLogger.Info("wrote model properties",
						slog.String("path", modelPropFilePath))
				}
			}
		}
	}
	return graphModels, nil
}

func recordsFileName(modelID string) string {
	return fmt.Sprintf("%s-records.json", modelID)
}

func (m *MetadataPreProcessor) WriteRecords(metadataDirectory string, datasetID string, graphModels []models.Model) error {
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
