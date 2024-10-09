package client

import "github.com/pennsieve/processor-pre-metadata/client/models/schema"

type Schema struct {
	modelNamesToSchemaElements      map[string]schema.Element
	linkedPropNamesToSchemaElements map[string]schema.Element
}

func NewSchema(schemaElements []schema.Element) *Schema {
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
	}
}

func (s *Schema) ModelCount() int {
	return len(s.modelNamesToSchemaElements)
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
