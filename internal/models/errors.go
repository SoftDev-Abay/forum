package models

import (
	"errors"
)

var ErrNoRecord = errors.New("models: no matching record found")

var ErrDuplicateEmail = errors.New("models: email already exists")


var ErrDuplicateUsername = errors.New("models: username already exists")
