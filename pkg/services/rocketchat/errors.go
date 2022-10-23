package rocketchat

import (
	"errors"
)

var (
	ErrAuthSessionExpired = errors.New("authorization session expired")
)
