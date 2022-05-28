package file_provider

import "io"

type Opener = func() (io.ReadCloser, error)

type FileForPath struct {
	Path   string
	Opener Opener
}
