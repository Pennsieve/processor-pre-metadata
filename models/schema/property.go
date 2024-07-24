package schema

type Property struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	DataType    string `json:"dataType"`
	Required    bool   `json:"required"`
	Index       int    `json:"index"`
}
