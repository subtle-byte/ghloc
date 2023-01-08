package loc_count

import "io"

type FilesLocCounter struct {
	fileLOCCounter fileLOCCounter
	locsForPaths   []LOCForPath
}

func NewFilesLOCCounter() *FilesLocCounter {
	return &FilesLocCounter{
		fileLOCCounter: newFileLOCCounter(),
	}
}

type LOCForPath struct {
	Path string
	LOC  int
}

func (c *FilesLocCounter) AddFile(path string, file io.Reader) error {
	loc, err := c.fileLOCCounter.Count(file)
	if err != nil {
		return err
	}
	c.locsForPaths = append(c.locsForPaths, LOCForPath{
		Path: path,
		LOC:  loc,
	})
	return nil
}

func (c *FilesLocCounter) GetLOCsForPaths() []LOCForPath {
	return c.locsForPaths
}
