package models

const ModelType = "concept"

// *Key must match the json struct tag for the property

const TypeKey = "type"
const IDKey = "id"
const NameKey = "name"
const DisplayNameKey = "displayName"

type Model struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
}

func ModelFromMap(jsonMap map[string]any) *Model {
	if jsonMap[TypeKey] != ModelType {
		return nil
	}
	return &Model{
		ID:          jsonMap[IDKey].(string),
		Type:        jsonMap[TypeKey].(string),
		Name:        jsonMap[NameKey].(string),
		DisplayName: jsonMap[DisplayNameKey].(string),
	}
}
