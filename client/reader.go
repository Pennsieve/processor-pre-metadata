package client

import (
	"encoding/json"
	"fmt"
	"github.com/pennsieve/processor-pre-metadata/client/models/instance"
	"github.com/pennsieve/processor-pre-metadata/client/models/schema"
	"github.com/pennsieve/processor-pre-metadata/client/paths"
	"os"
	"path/filepath"
	"strings"
)

// A Reader can be used to read the metadata records once they have been downloaded by the pre-processor
type Reader struct {
	MetadataDirectory               string
	ModelNamesToSchemaElements      map[string]schema.Element
	LinkedPropNamesToSchemaElements map[string]schema.Element
}

// NewReader returns a pointer to a new Reader instance. The rootDirectory argument should be
// the parent directory of the metadata directory.
func NewReader(rootDirectory string) (*Reader, error) {
	reader := Reader{
		MetadataDirectory:               filepath.Join(rootDirectory, paths.MetadataDirectory),
		ModelNamesToSchemaElements:      map[string]schema.Element{},
		LinkedPropNamesToSchemaElements: map[string]schema.Element{},
	}
	schemaFilePath := filepath.Join(reader.MetadataDirectory, paths.SchemaFilePath)
	schemaFile, err := os.Open(schemaFilePath)
	if err != nil {
		return nil, fmt.Errorf("error opening schema file %s: %w", schemaFilePath, err)
	}
	defer schemaFile.Close()

	var elements []schema.Element
	if err := json.NewDecoder(schemaFile).Decode(&elements); err != nil {
		return nil, fmt.Errorf("error decoding schema file %s: %w", schemaFilePath, err)
	}
	for _, e := range elements {
		if e.IsModel() {
			reader.ModelNamesToSchemaElements[e.Name] = e
		} else if e.IsLinkedProperty() {
			reader.LinkedPropNamesToSchemaElements[e.Name] = e
		}
	}
	return &reader, nil
}
func (r *Reader) GetRecordsForModel(modelName string) ([]instance.Record, error) {
	modelElement, isModel := r.ModelNamesToSchemaElements[strings.ToLower(modelName)]
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

func (r *Reader) GetLinkInstancesForProperty(linkedPropertyName string) ([]instance.LinkedProperty, error) {
	linkElement, isLink := r.LinkedPropNamesToSchemaElements[strings.ToLower(linkedPropertyName)]
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
