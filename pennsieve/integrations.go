package pennsieve

import (
	"encoding/json"
	"fmt"
	"github.com/pennsieve/processor-pre-metadata/models"
	"github.com/pennsieve/processor-pre-metadata/util"
	"io"
	"net/http"
)

func (s *Session) GetIntegration(integrationID string) (models.Integration, error) {
	url := fmt.Sprintf("%s/integrations/%s", s.API2Host, integrationID)

	res, err := s.InvokePennsieve(http.MethodGet, url, nil)
	if err != nil {
		return models.Integration{}, err
	}
	defer util.CloseAndWarn(res)

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return models.Integration{}, fmt.Errorf("error reading response from GET %s: %w", url, err)
	}

	var integration models.Integration
	if err := json.Unmarshal(body, &integration); err != nil {
		rawResponse := string(body)
		return models.Integration{}, fmt.Errorf(
			"error unmarshalling response [%s] from GET %s: %w",
			rawResponse,
			url,
			err)
	}

	return integration, nil
}
