package schema

const ProxyName = "belongs_to"
const ProxyDisplayName = "Belongs To"

// NullableRelationship represents an item in metadata/schema/relationships.json
// where unlike a usual Relationship, the special proxy relationship will have "to" and "from" null.
type NullableRelationship struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	DisplayName string  `json:"displayName"`
	From        *string `json:"from"`
	To          *string `json:"to"`
}

func (r NullableRelationship) IsProxy() bool {
	return IsProxy(r)
}

func IsProxy(r NullableRelationship) bool {
	return r.To == nil && r.From == nil && r.Name == ProxyName && r.DisplayName == ProxyDisplayName
}
