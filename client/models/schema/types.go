package schema

import (
	"fmt"
)

type Type string

const ModelType Type = "concept"
const RelationshipType Type = "schemaRelationship"
const LinkedPropertyType Type = "schemaLinkedProperty"

// TypeKey and other *Keys must match the json struct tag for the property
const TypeKey = "type"
const IDKey = "id"
const NameKey = "name"
const DisplayNameKey = "displayName"

type Element struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
}

func (e Element) isType(tipe string) bool {
	return e.Type == tipe
}

func (e Element) IsModel() bool {
	return e.isType(string(ModelType))
}

func (e Element) IsLinkedProperty() bool {
	return e.isType(string(LinkedPropertyType))
}

func ElementFromMap(jsonMap map[string]any) Element {
	return Element{
		ID:          jsonMap[IDKey].(string),
		Type:        jsonMap[TypeKey].(string),
		Name:        jsonMap[NameKey].(string),
		DisplayName: jsonMap[DisplayNameKey].(string),
	}
}

func IsType(jsonMap map[string]any, schemaType Type) (bool, error) {
	typeValAny, ok := jsonMap[TypeKey]
	if !ok {
		return false, fmt.Errorf("missing expected key %s: %s", TypeKey, jsonMap)
	}
	switch typeVal := typeValAny.(type) {
	case string:
		if typeVal == string(schemaType) {
			return true, nil
		}
		return false, nil
	default:
		return false, fmt.Errorf("expected string value at key %s in %s; got %T", TypeKey, jsonMap, typeVal)
	}

}

func IsModel(jsonMap map[string]any) (bool, error) {
	return IsType(jsonMap, ModelType)
}

func IsRelationship(jsonMap map[string]any) (bool, error) {
	return IsType(jsonMap, RelationshipType)
}

func IsLinkedProperty(jsonMap map[string]any) (bool, error) {
	return IsType(jsonMap, LinkedPropertyType)
}

func FromMap(jsonMap map[string]any) (any, error) {
	if model, err := ModelFromMap(jsonMap); err != nil {
		return nil, err
	} else if model != nil {
		return model, nil
	}
	if relationship, err := RelationshipFromMap(jsonMap); err != nil {
		return nil, err
	} else if relationship != nil {
		return relationship, nil
	}
	if linkedProp, err := LinkedPropertyFromMap(jsonMap); err != nil {
		return nil, err
	} else if linkedProp != nil {
		return linkedProp, nil
	}
	return nil, fmt.Errorf("unknown schema element type: %s", jsonMap)
}
