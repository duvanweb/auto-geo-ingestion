package errors

import (
	"net/http"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// ErrorResponse represents the body of an HTTP error response.
type ErrorResponse struct {
	Message string `json:"message"`
}

// WriteError writes an HTTP error response with the given status code and error message.
func WriteError(w http.ResponseWriter, statusCode int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	_ = json.NewEncoder(w).Encode(ErrorResponse{Message: err.Error()})
}
