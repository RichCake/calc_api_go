package storage

import "errors"

var (
	ErrItemNotFound  = errors.New("item not found")
	ErrUsernameTaken = errors.New("username is taken")
)
