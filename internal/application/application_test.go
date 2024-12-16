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
				t.Errorf("Wrong status code: expected %v got %v", http.StatusOK, res.StatusCode)
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