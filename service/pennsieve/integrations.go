package pennsieve

import (
	"encoding/json"
	"fmt"
	"github.com/pennsieve/processor-pre-metadata/service/util"
	"io"
	"net/http"
)

func (s *Session) GetIntegration(integrationID string) (Integration, error) {
	url := fmt.Sprintf("%s/integrations/%s", s.API2Host, integrationID)

	res, err := s.InvokePennsieve(http.MethodGet, url, nil)
	if err != nil {
		return Integration{}, err
	}
	defer util.CloseAndWarn(res)

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return Integration{}, fmt.Errorf("error reading response from GET %s: %w", url, err)
	}

	var integration Integration
	if err := json.Unmarshal(body, &integration); err != nil {
		rawResponse := string(body)
		return Integration{}, fmt.Errorf(
			"error unmarshalling response [%s] from GET %s: %w",
			rawResponse,
			url,
			err)
	}

	return integration, nil
}

type Integration struct {
	Uuid          string   `json:"uuid"`
	ApplicationID int64    `json:"applicationId"`
	DatasetNodeID string   `json:"datasetId"`
	PackageIDs    []string `json:"packageIds"`
	Params        any      `json:"params"`
}
