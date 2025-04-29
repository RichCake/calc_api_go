package auth

import "errors"

var (
	ErrUserExists     = errors.New("user already exists")
	ErrBadCredentials = errors.New("login or password is incorrect")
	ErrEncryption     = errors.New("unknown encryption error")
	ErrService        = errors.New("unknown service error")
	ErrStorage        = errors.New("unknown storage error")
)
