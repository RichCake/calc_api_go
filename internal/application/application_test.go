package application_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RichCake/calc_api_go/internal/application"
	"github.com/RichCake/calc_api_go/pkg/calculation"
)

func TestCalcHandlerSuccess(t *testing.T) {
	for _, tt := range calculation.ValidTestSet {
		t.Run(tt.Name, func(t *testing.T) {
			input, err := json.Marshal(map[string]string{"expression": tt.Expression})
			if err != nil {
				t.Fatalf("Error in input data '%v': %v", tt.Expression, err.Error())
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/calculate", bytes.NewReader(input))
			w := httptest.NewRecorder()
			handler := http.HandlerFunc(application.CalcHandler)
			handler.ServeHTTP(w, req)

			res := w.Result()
			data, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("Error while reading body: %v", err.Error())
			}
			defer res.Body.Close()

			if res.StatusCode != http.StatusOK {
				t.Errorf("Wrong status code: expected %v got %v", http.StatusOK, res.StatusCode)
			}

			var response map[string]float64
			err = json.Unmarshal(data, &response)
			if err != nil {
				t.Fatalf("Failed to parse response: %v", err)
			}

			answer, ok := response["result"]
			if !ok {
				t.Fatalf("Key 'result' was not found in '%v'", response)
			}
			if answer != tt.Expected_answer {
				t.Errorf("Wrong answer of expression '%v': expected %v got %v", tt.Expression, tt.Expected_answer, answer)
			}
		})
	}
}

func TestCalcHandlerUnprocessableEntity(t *testing.T) {
	for _, tt := range calculation.InvalidTestSet {
		t.Run(tt.Name, func(t *testing.T) {
			input, err := json.Marshal(map[string]string{"expression": tt.Expression})
			if err != nil {
				t.Fatalf("Error in input data '%v': %v", tt.Expression, err.Error())
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/calculate", bytes.NewReader(input))
			w := httptest.NewRecorder()
			handler := http.HandlerFunc(application.CalcHandler)
			handler.ServeHTTP(w, req)

			res := w.Result()
			data, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("Error while reading body: %v", err.Error())
			}
			defer res.Body.Close()

			if res.StatusCode != http.StatusUnprocessableEntity {
				t.Errorf("Wrong status code: expected %v got %v", http.StatusUnprocessableEntity, res.StatusCode)
			}

			var response map[string]string
			err = json.Unmarshal(data, &response)
			if err != nil {
				t.Fatalf("Failed to parse response: %v", err)
			}

			_, ok := response["error"]
			if !ok {
				t.Fatalf("Key 'result' was not found in '%v'", response)
			}
		})
	}
}

func TestCalcHandlerBadRequest(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		statusCode int
		errorMsg   string
	}{
		{"Empty body", "", http.StatusBadRequest, "missing request body"},
		{"Empty Expression", `{"expression":""}`, http.StatusBadRequest, "'expression' field is required"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/calculate", bytes.NewReader([]byte(tt.body)))
			w := httptest.NewRecorder()
			handler := http.HandlerFunc(application.CalcHandler)
			handler.ServeHTTP(w, req)

			res := w.Result()
			data, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("Error while reading body: %v", err.Error())
			}
			defer res.Body.Close()

			if res.StatusCode != tt.statusCode {
				t.Errorf("Wrong status code: expected %v got %v", tt.statusCode, res.StatusCode)
			}

			var response map[string]string
			err = json.Unmarshal(data, &response)
			if err != nil {
				t.Fatalf("Failed to parse response: %v", err)
			}

			errorMsg, ok := response["error"]
			if !ok {
				t.Fatalf("Key 'error' was not found in '%v'", response)
			}
			if errorMsg != tt.errorMsg {
				t.Errorf("Wrong error message: expected '%v' got '%v'", tt.errorMsg, errorMsg)
			}
		})
	}
}

func TestCalcHandlerMethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/calculate", nil)
	w := httptest.NewRecorder()
	handler := http.HandlerFunc(application.CalcHandler)
	handler.ServeHTTP(w, req)

	res := w.Result()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("Error while reading body: %v", err.Error())
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Wrong status code: expected %v got %v", http.StatusMethodNotAllowed, res.StatusCode)
	}

	var response map[string]string
	err = json.Unmarshal(data, &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	errorMsg, ok := response["error"]
	if !ok {
		t.Fatalf("Key 'error' was not found in '%v'", response)
	}
	if errorMsg != "method not allowed" {
		t.Errorf("Wrong error message: expected 'method not allowed' got '%v'", errorMsg)
	}
}