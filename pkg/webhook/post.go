package webhook

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

func Post(url string, m interface{}) error {
	//Encode the data
	postBody, err := json.Marshal(m)
	if err != nil {
		return errors.New(fmt.Sprintf("error marshalling request: %v", err))
	}
	responseBody := bytes.NewBuffer(postBody)

	//Leverage Go's HTTP Post function to make request
	_, err = http.Post(url, "application/json", responseBody)
	if err != nil {
		return errors.New(fmt.Sprintf("error when calling callback webhook %v", err))
	}
	return nil
}
