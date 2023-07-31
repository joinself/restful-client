package webhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func Post(url string, m interface{}) error {
	var postBody []byte
	var err error

	//Encode the data
	switch pb := m.(type) {
	case []byte:
		postBody = pb
	default:
		postBody, err = json.Marshal(m)
		if err != nil {
			return fmt.Errorf("error marshalling request: %v", err)
		}
	}

	responseBody := bytes.NewBuffer(postBody)

	//Leverage Go's HTTP Post function to make request
	_, err = http.Post(url, "application/json", responseBody)
	if err != nil {
		return fmt.Errorf("error when calling callback webhook %v", err)
	}
	return nil
}
