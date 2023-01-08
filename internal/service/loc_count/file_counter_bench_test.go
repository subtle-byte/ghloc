package loc_count

import (
	"bytes"
	"io"
	"os"
	"testing"
)

const locInLongText = 8471

func BenchmarkFileLOCCounter(b *testing.B) {
	file, err := os.Open("./testdata/long_text.txt")
	if err != nil {
		b.Fatal(err)
	}
	zip, err := io.ReadAll(file)
	if err != nil {
		b.Fatal(err)
	}
	locCounter := newFileLOCCounter()
	for i := 0; i < b.N; i++ {
		r := bytes.NewReader(zip)
		loc, err := locCounter.Count(r)
		if err != nil {
			b.Fatal(err)
		}
		if loc == 0 {
			b.Fatal("loc must be > 0")
		}
		if locInLongText != loc {
			b.Fatalf("locInLongText(%v) != loc(%v)", locInLongText, loc)
		}
	}
}
