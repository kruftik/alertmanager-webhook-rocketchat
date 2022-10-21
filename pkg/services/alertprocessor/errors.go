package alertprocessor

import (
	"errors"
)

var (
	ErrDefaultChannelNotDefined = errors.New("default Rocket.Chat channel name is not defined in configuration")
)
