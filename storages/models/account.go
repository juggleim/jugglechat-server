package models

import "errors"

var (
	ErrAccountConflict     = errors.New("account already in use")
	ErrAccountUserNotFound = errors.New("account user not found")
	ErrAccountAppNotFound  = errors.New("account app not found")
)
