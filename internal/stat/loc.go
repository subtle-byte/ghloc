package stat

import "io"

type LocCounter struct {
	fileLOCCounter fileLOCCounter
	locsForPaths   []LOCForPath
}

func NewLOCCounter() *LocCounter {
	return &LocCounter{
		fileLOCCounter: newFileLOCCounter(),
	}
}

type LOCForPath struct {
	Path string
	LOC  int
}

func (c *LocCounter) AddFile(path string, file io.Reader) error {
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

func (c *LocCounter) GetLOCsForPaths() []LOCForPath {
	return c.locsForPaths
}
