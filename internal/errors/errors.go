package errors

import (
	"fmt"
)

var (
	NotFound       = fmt.Errorf("not found")
	NoSpaceLeft    = fmt.Errorf("no space left")
	NotImplemented = fmt.Errorf("not implemented")
)
