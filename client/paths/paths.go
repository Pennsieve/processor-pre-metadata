package paths

import (
	"fmt"
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

// MetadataDirectory is the directory all metadata info will be placed in relative to the input directory
const MetadataDirectory = "metadata"

// SchemaDirectory is the directory schema elements will be placed in relative to the metadata directory
const SchemaDirectory = "schema"

// PropertiesDirectory is the directory property files will be placed in relative to the schema directory
const PropertiesDirectory = "properties"

// InstancesDirectory is the directory instance info will be placed in relative to the metadata directory
const InstancesDirectory = "instances"

// RecordsDirectory is the directory record files will be placed in relative to the instances directory
const RecordsDirectory = "records"

// RelationshipsDirectory is the directory relationship files will be placed in relative to the instances directory
const RelationshipsDirectory = "relationships"

// LinkedPropertiesDirectory is the directory linked properties files will be placed in relative to the instances directory
const LinkedPropertiesDirectory = "linkedProperties"

// SchemaFilePath is the path to the schema json file relative to the metadata directory
var SchemaFilePath = filepath.Join(SchemaDirectory, "graphSchema.json")

// PropertiesFilePath the path of the properties file for the given model relative to the metadata directory
func PropertiesFilePath(modelID string) string {
	return filepath.Join(SchemaDirectory, PropertiesDirectory, fmt.Sprintf("%s.json", modelID))
}

// RecordsFilePath the path of the records file for the given model relative to the metadata directory
func RecordsFilePath(modelID string) string {
	return filepath.Join(InstancesDirectory, RecordsDirectory, fmt.Sprintf("%s.json", modelID))
}

// RelationshipInstancesFilePath the path of the instances file for the given schema relationship relative to the metadata directory
func RelationshipInstancesFilePath(schemaRelationshipID string) string {
	return filepath.Join(InstancesDirectory, RelationshipsDirectory, fmt.Sprintf("%s.json", schemaRelationshipID))
}

// LinkedPropertyInstancesFilePath the path of the instances file for the given schema linked property relative to the metadata directory
func LinkedPropertyInstancesFilePath(schemaLinkedPropertyID string) string {
	return filepath.Join(InstancesDirectory, LinkedPropertiesDirectory, fmt.Sprintf("%s.json", schemaLinkedPropertyID))
}
