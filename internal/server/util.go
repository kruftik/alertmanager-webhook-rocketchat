package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/prometheus/alertmanager/template"
)

// JSONResponse is the webhook http response
type JSONResponse struct {
	Status  int
	Message string
}

func sendJSONResponse(w http.ResponseWriter, status int, message string) error {
	data := JSONResponse{
		Status:  status,
		Message: message,
	}

	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		return fmt.Errorf("cannot marshal JSON: %w", err)
	}

	return nil
}

func readRequestBody(r *http.Request) (template.Data, error) {
	// Do not forget to close the body at the end
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(r.Body)

	// Extract data from the body in the Data template provided by AlertManager
	data := template.Data{}
	err := json.NewDecoder(r.Body).Decode(&data)

	return data, err
}
