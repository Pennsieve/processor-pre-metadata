package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/pennsieve/processor-pre-metadata/client/models/instance"
	"github.com/pennsieve/processor-pre-metadata/client/models/schema"
	"github.com/pennsieve/processor-pre-metadata/client/paths"
	"os"
	"path/filepath"
	"slices"
	"strings"
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

// GetProxiesForModel returns the proxy instances for the given model, grouped by the record IDs to which the proxies are linked.
// That is, it returns a map from record ID to a slice of proxy instances for that record. Each proxy instance
// represents one package that is linked to the record
func (r *Reader) GetProxiesForModel(modelName string) (map[string][]instance.Proxy, error) {
	model, isModel := r.Schema.ModelByName(modelName)
	if !isModel {
		return nil, fmt.Errorf("model %s not found", modelName)
	}
	var proxiesByRecordID = make(map[string][]instance.Proxy)
	if r.Schema.Proxy() == nil {
		// If there is no Proxy schema, there should be no proxy instances
		return proxiesByRecordID, nil
	}
	parentDirectoryPath := filepath.Join(r.MetadataDirectory, paths.ProxyInstancesForModelDirectory(model.ID))
	dirEntries, err := os.ReadDir(parentDirectoryPath)
	if err != nil {
		// if the directory does not exist, this is not an error. Just means no proxy instances for this model
		if errors.Is(err, os.ErrNotExist) {
			return proxiesByRecordID, nil
		} else {
			return nil, fmt.Errorf("error reading proxy instance directory %s for model %s: %w", parentDirectoryPath, modelName, err)
		}
	}
	for _, dirEntry := range dirEntries {
		if recordID, isInstanceFile := getProxyRecordID(dirEntry); isInstanceFile {
			instanceFilePath := filepath.Join(parentDirectoryPath, dirEntry.Name())
			proxyInstances, err := readProxyInstanceFile(instanceFilePath)
			if err != nil {
				return nil, err
			}
			proxiesByRecordID[recordID] = proxyInstances
		}
	}
	return proxiesByRecordID, nil

}

func getProxyRecordID(dirEntry os.DirEntry) (string, bool) {
	extension := filepath.Ext(dirEntry.Name())
	if dirEntry.Type().IsRegular() && extension == ".json" {
		return strings.TrimSuffix(dirEntry.Name(), extension), true
	}
	return "", false
}

func readProxyInstanceFile(proxyInstanceFilePath string) ([]instance.Proxy, error) {
	proxyInstanceFile, err := os.Open(proxyInstanceFilePath)
	if err != nil {
		return nil, fmt.Errorf("error opening proxy instance file %s: %w", proxyInstanceFilePath, err)
	}
	defer proxyInstanceFile.Close()

	var rawInstances []instance.RawFromFile
	if err := json.NewDecoder(proxyInstanceFile).Decode(&rawInstances); err != nil {
		return nil, fmt.Errorf("error decoding proxy instances from file %s: %w", proxyInstanceFilePath, err)
	}
	var proxies []instance.Proxy
	for _, raw := range rawInstances {
		var proxyID instance.ProxyID
		if err := json.Unmarshal(raw[0], &proxyID); err != nil {
			return nil, fmt.Errorf("error decoding ProxyID %s from proxy instance file %s: %w", raw[0], proxyInstanceFilePath, err)
		}
		var proxyPackage instance.ProxyPackage
		if err := json.Unmarshal(raw[1], &proxyPackage); err != nil {
			return nil, fmt.Errorf("error decoding ProxyPackage %s from proxy instance file %s: %w", raw[1], proxyInstanceFilePath, err)
		}
		proxies = append(proxies, instance.Proxy{
			ProxyID:      proxyID,
			ProxyPackage: proxyPackage,
		})
	}

	return proxies, nil
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
