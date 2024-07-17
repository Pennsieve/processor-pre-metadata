package pennsieve

import (
	"encoding/json"
	"fmt"
	"github.com/pennsieve/processor-pre-metadata/util"
	"net/http"
)

func (s *Session) GetRecordsPage(datasetID string, modelID string, limit int, offset int) ([]map[string]any, error) {
	url := fmt.Sprintf("%s/models/v1/datasets/%s/concepts/%s/instances?limit=%d&offset=%d", s.APIHost, datasetID, modelID, limit, offset)
	res, err := s.InvokePennsieve(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	defer util.CloseAndWarn(res)

	decoder := json.NewDecoder(res.Body)

	var batch []map[string]any
	if err = decoder.Decode(&batch); err != nil {
		return nil, fmt.Errorf("error decoding records for model %s: %w", modelID, err)
	}
	return batch, nil
}

func (s *Session) GetAllRecords(datasetID string, modelID string, batchSize int) ([]map[string]any, error) {
	if batchSize <= 0 {
		return nil, fmt.Errorf("illegal batchSize; must be > 0: %d", batchSize)
	}

	var records []map[string]any
	for offset := 0; true; {
		if batch, err := s.GetRecordsPage(datasetID, modelID, batchSize, offset); err != nil {
			return nil, err
		} else {
			records = append(records, batch...)
			if len(batch) < batchSize {
				// this endpoint does not tell use when it's returned the final page, so we have to call it until it returns empty
				// but also, if it has returned less than batchSize, then there should not be any more records.
				break
			} else {
				offset = offset + len(batch)
			}
		}
	}
	return records, nil
}
