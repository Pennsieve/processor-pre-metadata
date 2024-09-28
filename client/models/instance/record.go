package instance

import "time"

type Record struct {
	CreatedAt time.Time  `json:"createdAt"`
	CreatedBy string     `json:"createdBy"`
	ID        string     `json:"id"`
	Type      string     `json:"type"`
	UpdatedAt time.Time  `json:"updatedAt"`
	UpdatedBy string     `json:"updatedBy"`
	Values    []Property `json:"values"`
}
