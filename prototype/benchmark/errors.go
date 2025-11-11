package benchmark

import "errors"

var (
	ErrSpaceFull = errors.New("space full")
	ErrEOF       = errors.New("EOF")
)
