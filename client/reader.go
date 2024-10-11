package client

import (
	"encoding/json"
	"fmt"
	"github.com/pennsieve/processor-pre-metadata/client/models/instance"
	"github.com/pennsieve/processor-pre-metadata/client/models/schema"
	"github.com/pennsieve/processor-pre-metadata/client/paths"
	"os"
	"path/filepath"
	"slices"
)

// A Reader can be used to read the metadata records once they have been downloaded by the pre-processor
type Reader struct {
	MetadataDirectory string
	Schema            *Schema
}

// NewReader returns a pointer to a new Reader instance. The rootDirectory argument should be
// the parent directory of the metadata directory.
func NewReader(rootDirectory string) (*Reader, error) {
	reader := Reader{
		MetadataDirectory: filepath.Join(rootDirectory, paths.MetadataDirectory),
	}
	var proxy *schema.NullableRelationship
	relationshipsFilePath := filepath.Join(reader.MetadataDirectory, paths.RelationshipSchemasFilePath)
	var relationships []schema.NullableRelationship
	if err := readJsonFile(relationshipsFilePath, &relationships); err != nil {
		return nil, err
	}
	proxyIndex := slices.IndexFunc(relationships, schema.IsProxy)
	if proxyIndex != -1 {
		proxy = &relationships[proxyIndex]
	}
	schemaFilePath := filepath.Join(reader.MetadataDirectory, paths.SchemaFilePath)
	var elements []schema.Element
	if err := readJsonFile(schemaFilePath, &elements); err != nil {
		return nil, err
	}
	reader.Schema = NewSchema(elements, proxy)
	return &reader, nil
}
func (r *Reader) GetRecordsForModel(modelName string) ([]instance.Record, error) {
	modelElement, isModel := r.Schema.ModelByName(modelName)
	if !isModel {
		return nil, fmt.Errorf("model %s not found", modelName)
	}
	recordsFilePath := filepath.Join(r.MetadataDirectory, paths.RecordsFilePath(modelElement.ID))
	recordsFile, err := os.Open(recordsFilePath)
	if err != nil {
		return nil, fmt.Errorf("error opening records file %s for %s: %w", recordsFilePath, modelName, err)
	}
	defer recordsFile.Close()

	var records []instance.Record
	if err := json.NewDecoder(recordsFile).Decode(&records); err != nil {
		return nil, fmt.Errorf("error decoding records file %s for %s: %w", recordsFilePath, modelName, err)
	}
	return records, nil
}

func (r *Reader) GetProxiesForModel(modelName string) (map[string][]instance.Proxy, error) {
	_, isModel := r.Schema.ModelByName(modelName)
	if !isModel {
		return nil, fmt.Errorf("model %s not found", modelName)
	}
	var proxiesByRecordID = make(map[string][]instance.Proxy)
	if r.Schema.Proxy() == nil {
		// If there is no Proxy schema, there should be no proxy instances
		return proxiesByRecordID, nil
	}
	return proxiesByRecordID, nil

}

func (r *Reader) GetLinkInstancesForProperty(linkedPropertyName string) ([]instance.LinkedProperty, error) {
	linkElement, isLink := r.Schema.LinkedPropertyByName(linkedPropertyName)
	if !isLink {
		return nil, fmt.Errorf("linked property %s not found", linkedPropertyName)
	}
	linksFilePath := filepath.Join(r.MetadataDirectory, paths.LinkedPropertyInstancesFilePath(linkElement.ID))
	linksFile, err := os.Open(linksFilePath)
	if err != nil {
		return nil, fmt.Errorf("error opening linked properties instance file %s for %s: %w",
			linksFilePath,
			linkedPropertyName,
			err)
	}
	defer linksFile.Close()

	var links []instance.LinkedProperty
	if err := json.NewDecoder(linksFile).Decode(&links); err != nil {
		return nil, fmt.Errorf("error decoding linked properties instance file %s for %s: %w",
			linksFilePath,
			linkedPropertyName,
			err)
	}
	return links, nil

}

func readJsonFile(filePath string, value any) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("error opening file %s: %w", filePath, err)
	}
	defer file.Close()

	if err := json.NewDecoder(file).Decode(value); err != nil {
		return fmt.Errorf("error decoding file %s: %w", filePath, err)
	}
	return nil
}
