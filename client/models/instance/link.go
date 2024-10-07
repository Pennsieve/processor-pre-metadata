package instance

type LinkedProperty struct {
	DisplayName          string `json:"displayName"`
	From                 string `json:"from"`
	Id                   string `json:"id"`
	Name                 string `json:"name"`
	SchemaRelationshipId string `json:"schemaRelationshipId"`
	To                   string `json:"to"`
	Type                 string `json:"type"`
}
