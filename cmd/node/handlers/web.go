// Package handlers
package handlers

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
)

func Decode[T any](r *http.Request) (T, error) {
	var v T

	switch r.Header.Get("Content-Type") {
	case "application/json":
		if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
			return v, err
		}
	case "application/xml":
		if err := xml.NewDecoder(r.Body).Decode(&v); err != nil {
			return v, err
		}
	}

	return v, nil
}

func Encode(w http.ResponseWriter, r *http.Request, v any) error {
	mimeType := r.Header.Get("Accept")

	w.Header().Set("Content-Type", mimeType)
	w.WriteHeader(http.StatusOK)

	switch mimeType {
	case "application/json":
		_ = json.NewEncoder(w).Encode(v)
	case "application/xml":
		_ = xml.NewEncoder(w).Encode(v)
	}

	return nil
}
