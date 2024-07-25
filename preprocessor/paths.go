package preprocessor

import (
	"fmt"
	"os"
	"path/filepath"
)

// layout relative to input directory:
// metadata/
// ├── schema/
// │   ├── graphSchema.json
// │   └── properties/
// │       ├── <model-id-1>.json
// │       └── <model-id-2>.json
// └── instances/
//     ├── records/
//     │   ├── <model-id-1>.json
//     │   └── <model-id-2>.json
//     ├── relationships/
//     │   ├── <schemaRelationship-id-1>.json
//     │   ├── <schemaRelationship-id-2>.json
//     │   └── <schemaRelationship-id-3>.json
//     └── linkedProperties/
//         └── <schemaLinkedProperty-id-1>.json

// metadataDirectory is the directory all metadata info will be placed in relative to the input directory
const metadataDirectory = "metadata"

// schemaDirectory is the directory schema elements will be placed in relative to the metadata directory
const schemaDirectory = "schema"

// propertiesDirectory is the directory property files will be placed in relative to the schema directory
const propertiesDirectory = "properties"

// instancesDirectory is the directory instance info will be placed in relative to the metadata directory
const instancesDirectory = "instances"

// recordsDirectory is the directory record files will be placed in relative to the instances directory
const recordsDirectory = "records"

// relationshipsDirectory is the directory relationship files will be placed in relative to the instances directory
const relationshipsDirectory = "relationships"

// linkedPropertiesDirectory is the directory linked properties files will be placed in relative to the instances directory
const linkedPropertiesDirectory = "linkedProperties"

// schemaFilePath is the path to the schema json file relative to the metadata directory
var schemaFilePath = filepath.Join(schemaDirectory, "graphSchema.json")

// MetadataPath gives the path to metadataDirectory relative to the input directory
func (m *MetadataPreProcessor) MetadataPath() string {
	return filepath.Join(m.InputDirectory, metadataDirectory)
}

// PropertiesPath gives the path to propertiesDirectory relative to the input directory
func (m *MetadataPreProcessor) PropertiesPath() string {
	return filepath.Join(m.MetadataPath(), schemaDirectory, propertiesDirectory)
}

// RecordsPath gives the path to recordsDirectory relative to the input directory
func (m *MetadataPreProcessor) RecordsPath() string {
	return filepath.Join(m.MetadataPath(), instancesDirectory, recordsDirectory)
}

// RelationshipsPath gives the path to relationshipsDirectory relative to the input directory
func (m *MetadataPreProcessor) RelationshipsPath() string {
	return filepath.Join(m.MetadataPath(), instancesDirectory, relationshipsDirectory)
}

// LinkedPropertiesPath gives the path to linkedPropertiesDirectory relative to the input directory
func (m *MetadataPreProcessor) LinkedPropertiesPath() string {
	return filepath.Join(m.MetadataPath(), instancesDirectory, linkedPropertiesDirectory)
}

// propertiesFilePath the path of the properties file for the given model relative to the metadata directory
func propertiesFilePath(modelID string) string {
	return filepath.Join(schemaDirectory, propertiesDirectory, fmt.Sprintf("%s.json", modelID))
}

// recordsFilePath the path of the records file for the given model relative to the metadata directory
func recordsFilePath(modelID string) string {
	return filepath.Join(instancesDirectory, recordsDirectory, fmt.Sprintf("%s.json", modelID))
}

// relationshipInstancesFilePath the path of the instances file for the given schema relationship relative to the metadata directory
func relationshipInstancesFilePath(schemaRelationshipID string) string {
	return filepath.Join(instancesDirectory, relationshipsDirectory, fmt.Sprintf("%s.json", schemaRelationshipID))
}

// linkedPropertyInstancesFilePath the path of the instances file for the given schema linked property relative to the metadata directory
func linkedPropertyInstancesFilePath(schemaLinkedPropertyID string) string {
	return filepath.Join(instancesDirectory, linkedPropertiesDirectory, fmt.Sprintf("%s.json", schemaLinkedPropertyID))
}

func (m *MetadataPreProcessor) MkDirectories() error {
	leafDirectories := []string{
		m.PropertiesPath(),
		m.RecordsPath(),
		m.RelationshipsPath(),
		m.LinkedPropertiesPath(),
	}
	for _, leaf := range leafDirectories {
		if err := os.MkdirAll(leaf, 0755); err != nil {
			return fmt.Errorf("error creating directory %s: %w", leaf, err)
		}
	}
	return nil
}
