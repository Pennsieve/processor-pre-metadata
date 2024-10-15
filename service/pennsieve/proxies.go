package pennsieve

import (
	"encoding/json"
	"fmt"
	"github.com/pennsieve/processor-pre-metadata/service/util"
	"net/http"
)

// GetProxyInstancesForRecord returns an []any because we are only dumping result to a file if there are any
// proxies. So all we care about here is if the slice is empty or not.
func (s *Session) GetProxyInstancesForRecord(datasetID, modelID, recordID string) ([]any, error) {
	url := fmt.Sprintf("%s/models/datasets/%s/concepts/%s/instances/%s/files", s.APIHost, datasetID, modelID, recordID)
	res, err := s.InvokePennsieve(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	defer util.CloseAndWarn(res)

	var proxies []any
	if err := json.NewDecoder(res.Body).Decode(&proxies); err != nil {
		return nil, fmt.Errorf("error decoding proxies for record %s: %w", recordID, err)
	}
	return proxies, nil
}
