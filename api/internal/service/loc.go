package service

import (
	"io"
)

type locCounter struct {
	buf []byte
}

func newLOCCounter() locCounter {
	return locCounter{make([]byte, 1000)}
}

func (lc *locCounter) Count(r io.Reader) (int, error) {
	loc := 0
	nonSpace := false
	for {
		n, err := r.Read(lc.buf)
		for _, c := range lc.buf[:n] {
			if c > ' ' {
				nonSpace = true
			} else if c == '\n' {
				if nonSpace {
					loc++
				}
				nonSpace = false
			} else if c != ' ' && c != '\t' && c != '\r' {
				return 1, nil
			}
		}
		if err != nil {
			if err == io.EOF {
				if nonSpace {
					loc++
				}
				break
			}
			return 0, err
		}
	}

	return loc, nil
}
