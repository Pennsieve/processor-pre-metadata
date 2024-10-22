package client

import "github.com/pennsieve/processor-pre-metadata/client/models/schema"

type Schema struct {
	modelNamesToSchemaElements      map[string]schema.Element
	linkedPropNamesToSchemaElements map[string]schema.Element
	proxy                           *schema.NullableRelationship
}

func NewSchema(schemaElements []schema.Element, proxy *schema.NullableRelationship) *Schema {
	modelMap := make(map[string]schema.Element)
	linkMap := make(map[string]schema.Element)
	for _, e := range schemaElements {
		if e.IsModel() {
			modelMap[e.Name] = e
		} else if e.IsLinkedProperty() {
			linkMap[e.Name] = e
		}
	}
	return &Schema{
		modelNamesToSchemaElements:      modelMap,
		linkedPropNamesToSchemaElements: linkMap,
		proxy:                           proxy,
	}
}

func (s *Schema) ModelCount() int {
	return len(s.modelNamesToSchemaElements)
}

// ModelIDsByName returns a map model name -> model schema id
func (s *Schema) ModelIDsByName() map[string]string {
	nameToId := make(map[string]string, len(s.modelNamesToSchemaElements))
	for name, element := range s.modelNamesToSchemaElements {
		nameToId[name] = element.ID
	}
	return nameToId
}

func (s *Schema) ModelByName(modelName string) (model schema.Element, modelExists bool) {
	model, modelExists = s.modelNamesToSchemaElements[modelName]
	return
}

func (s *Schema) LinkedPropertyCount() int {
	return len(s.linkedPropNamesToSchemaElements)
}

func (s *Schema) LinkedPropertyByName(linkName string) (linkedProperty schema.Element, linkedPropertyExists bool) {
	linkedProperty, linkedPropertyExists = s.linkedPropNamesToSchemaElements[linkName]
	return
}

func (s *Schema) Proxy() *schema.NullableRelationship {
	return s.proxy
}
