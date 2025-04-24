package expression

import "errors"

var (
	ErrPendingTaskNotFount = errors.New("no pending task available")
	ErrExpressionNotFound  = errors.New("expression not found")
	ErrZeroDivisionTask    = errors.New("task with zero division")
	ErrTaskNotFound        = errors.New("task not found")
	ErrStorage             = errors.New("unknown error in storage")
	ErrService             = errors.New("unknown error in service")
)
