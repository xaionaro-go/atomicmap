package atomicmap

import (
	"github.com/xaionaro-go/atomicmap/errors"
)

var (
	NotFound        = errors.NotFound
	NoSpaceLeft     = errors.NoSpaceLeft
	NotImplemented  = errors.NotImplemented
	AlreadyGrowing  = errors.AlreadyGrowing
	ForbiddenToGrow = errors.ForbiddenToGrow
)
