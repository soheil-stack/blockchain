package core

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

func Send(method string, url string, dataSend, dataReceive any) error {
	var req *http.Request

	switch {
	case dataSend != nil:
		data, err := json.Marshal(dataSend)
		if err != nil {
			return err
		}
		req, err = http.NewRequest(method, url, bytes.NewReader(data))
		if err != nil {
			return nil
		}
	default:
		var err error
		req, err = http.NewRequest(method, url, nil)
		if err != nil {
			return err
		}
	}

	var client http.Client
	response, err := client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusNoContent {
		return nil
	}

	if response.StatusCode != http.StatusOK {
		msg, err := io.ReadAll(response.Body)
		if err != nil {
			return nil
		}
		return errors.New(string(msg))
	}

	if dataReceive != nil {
		if err := json.NewDecoder(response.Body).Decode(dataReceive); err != nil {
			return nil
		}
	}

	return nil
}
