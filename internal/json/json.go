package json

import (
	"encoding/json"
	"io"
	"net/http"
)

func Write(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func Read(r *http.Request, data any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(data)
}

// Add this to handle any generic reader (like Google's API response body)
func ReadFrom(reader io.Reader, data any) error {
	decoder := json.NewDecoder(reader)
	// decoder.DisallowUnknownFields()
	return decoder.Decode(data)
}
