package models

import (
	"time"

	"github.com/RichCake/calc_api_go/internal/services/calculation"
)

type Expression struct {
	ID int `json:"id"`
	Status string `json:"status"`
	Result float64 `json:"result"`
	BinaryTree *calculation.Tree
}

type Task struct {
	ID int
	ExpressionID int
	Status string
	Arg1 float64
	Arg2 float64
	Operation string
	OperationTime time.Duration
}