package preprocessor

import (
	"fmt"
	"github.com/pennsieve/processor-pre-metadata/client/paths"
	"os"
	"path/filepath"
)

// MetadataPath gives the path to metadataDirectory relative to the input directory
func (m *MetadataPreProcessor) MetadataPath() string {
	return filepath.Join(m.InputDirectory, paths.MetadataDirectory)
}

// PropertiesPath gives the path to propertiesDirectory relative to the input directory
func (m *MetadataPreProcessor) PropertiesPath() string {
	return filepath.Join(m.MetadataPath(), paths.SchemaDirectory, paths.PropertiesDirectory)
}

// RecordsPath gives the path to recordsDirectory relative to the input directory
func (m *MetadataPreProcessor) RecordsPath() string {
	return filepath.Join(m.MetadataPath(), paths.InstancesDirectory, paths.RecordsDirectory)
}

// RelationshipsPath gives the path to relationshipsDirectory relative to the input directory
func (m *MetadataPreProcessor) RelationshipsPath() string {
	return filepath.Join(m.MetadataPath(), paths.InstancesDirectory, paths.RelationshipsDirectory)
}

// LinkedPropertiesPath gives the path to linkedPropertiesDirectory relative to the input directory
func (m *MetadataPreProcessor) LinkedPropertiesPath() string {
	return filepath.Join(m.MetadataPath(), paths.InstancesDirectory, paths.LinkedPropertiesDirectory)
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
