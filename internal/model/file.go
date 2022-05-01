package model

import (
	"io"
)

type FileName = string

type Opener = func() (io.ReadCloser, error)
