package pennsieve

import (
	"fmt"
	"net/http"
)

func (s *Session) GetGraphSchema(datasetID string) (*http.Response, error) {
	url := fmt.Sprintf("%s/models/datasets/%s/concepts/schema/graph", s.APIHost, datasetID)

	return s.InvokePennsieve(http.MethodGet, url, nil)
}

func (s *Session) GetProperties(datasetID, modelID string) (*http.Response, error) {
	url := fmt.Sprintf("%s/models/datasets/%s/concepts/%s/properties", s.APIHost, datasetID, modelID)
	return s.InvokePennsieve(http.MethodGet, url, nil)
}
