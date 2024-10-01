package schema

import "encoding/json"

type Property struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	// DataType may be a string or may be a JSON object
	DataType json.RawMessage `json:"dataType"`
	Required bool            `json:"required"`
	Index    int             `json:"index"`
}
