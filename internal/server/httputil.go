package server

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"log/slog"
	"net/http"
)

func decode[T any](r *http.Request) (T, error) {
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
	default:
		return v, errors.New("unsupported content type")
	}

	return v, nil
}

func encode(w http.ResponseWriter, r *http.Request, v any) error {
	mimeType := r.Header.Get("Accept")

	switch mimeType {
	case "application/json", "application/xml":
	default:
		mimeType = "application/json"
	}

	if v == nil {
		w.WriteHeader(http.StatusNoContent)
		return nil
	}

	buf := new(bytes.Buffer)
	switch mimeType {
	case "application/json":
		if err := json.NewEncoder(buf).Encode(v); err != nil {
			return err
		}
	case "application/xml":
		if err := xml.NewEncoder(buf).Encode(v); err != nil {
			return err
		}
	}

	w.Header().Set("Content-Type", mimeType)
	w.WriteHeader(http.StatusOK)
	if _, err := buf.WriteTo(w); err != nil {
		slog.Error("failed to write response", "err", err, "path", r.URL.Path)
	}

	return nil
}
