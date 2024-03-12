package self

import (
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/joinself/restful-client/internal/entity"
)

type metrc struct {
	UUID      int      `json:"uuid"`
	Recipient string   `json:"recipient"`
	Actions   []string `json:"actions"`
	Date      int64    `json:"date`
}

type attestedFacts struct {
	Transactions []metrc
}

type transactionFact struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type transactionFacts struct {
	Facts []transactionFact `json:"facts"`
}

func parseIncomingMetrics(payload map[string]interface{}) ([]*entity.Metric, error) {
	metrics := []*entity.Metric{}
	for _, a := range payload["attestations"].([]interface{}) {
		p1 := a.(map[string]interface{})["payload"].(string)
		p, err := base64.RawURLEncoding.DecodeString(p1)
		if err != nil {
			return metrics, err
		}

		var x transactionFacts
		err = json.Unmarshal(p, &x)
		if err != nil {
			return metrics, err
		}
		for _, f := range x.Facts {
			var af attestedFacts
			err := json.Unmarshal([]byte(f.Value), &af)
			if err != nil {
				return metrics, err
			}
			for _, t := range af.Transactions {
				// Check if metric entry already exist...
				a, err := json.Marshal(t.Actions)
				if err != nil {
					return metrics, err
				}
				metrics = append(metrics, &entity.Metric{
					UUID:      t.UUID,
					Recipient: t.Recipient,
					Actions:   string(a),
					CreatedAt: time.Unix(t.Date, 0),
				})
			}

		}
	}
	return metrics, nil
}
