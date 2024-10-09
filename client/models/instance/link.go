package instance

type LinkedProperty struct {
	DisplayName          string `json:"displayName"`
	From                 string `json:"from"`
	ID                   string `json:"id"`
	Name                 string `json:"name"`
	SchemaRelationshipID string `json:"schemaRelationshipId"`
	To                   string `json:"to"`
	Type                 string `json:"type"`
}
