package calculation_test

import (
	"errors"
	"testing"

	"github.com/RichCake/calc_api_go/pkg/calculation"
)

func TestCalcValid(t *testing.T) {
	for _, tt := range calculation.ValidTestSet {
		t.Run(tt.Name, func(t *testing.T) {
			answer, err := calculation.Calc(tt.Expression)
			if err != nil {
				t.Fatalf("Error in expression (%v): %v", tt.Expression, err.Error())
			}
			if answer != tt.Expected_answer {
				t.Errorf("Wrong answer in expression '%v': expected '%v' got '%v'", tt.Expression, tt.Expected_answer, answer)
			}
		})
	}
}

func TestCalcInvalid(t *testing.T) {
	for _, tt := range calculation.InvalidTestSet {
		t.Run(tt.Name, func(t *testing.T) {
			_, err := calculation.Calc(tt.Expression)
			if !errors.Is(err, tt.Expected_error) {
				t.Errorf("Wrong error in expression '%v': expected '%v' got '%v'",tt.Expression, tt.Expected_error, err)
			}
		})
	}
}