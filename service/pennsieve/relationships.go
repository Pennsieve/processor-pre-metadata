package pennsieve

import (
	"fmt"
	"net/http"
)

func (s *Session) GetRelationshipInstances(datasetID, schemaRelationshipID string) (*http.Response, error) {
	url := fmt.Sprintf("%s/models/v1/datasets/%s/relationships/%s/instances", s.APIHost, datasetID, schemaRelationshipID)
	return s.InvokePennsieve(http.MethodGet, url, nil)
}

func (s *Session) GetRelationshipSchemas(datasetID string) (*http.Response, error) {
	url := fmt.Sprintf("%s/models/datasets/%s/relationships", s.APIHost, datasetID)
	return s.InvokePennsieve(http.MethodGet, url, nil)
}
