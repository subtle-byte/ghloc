package rest

import (
	"encoding/json"
	"io"
	"sort"

	"github.com/subtle-byte/ghloc/internal/service/loc_count"
)

type SortedStat loc_count.StatTree

func (st *SortedStat) MarshalJSON() ([]byte, error) {
	w := &Buffer{}
	err := st.marshalJson(w)
	return w.Bytes(), err
}

func (st *SortedStat) marshalJson(w *Buffer) error {
	var err error
	keepErr := func(err2 error) {
		if err == nil {
			err = err2
		}
	}
	encode := func(v interface{}) {
		keepErr(json.NewEncoder(w).Encode(v))
		w.UnwriteByte() // remove newline inserted by json.Encoder
	}

	if len(st.Children) == 0 {
		encode(st.LOC)
		return err
	}

	w.WriteString(`{"loc":`)
	encode(st.LOC)

	if len(st.LOCByLangs) != 0 {
		w.WriteString(`,"locByLangs":`)
		locByLangs := mapItems(st.LOCByLangs)
		sort.Slice(locByLangs, func(i, j int) bool {
			return locByLangs[i].Value > locByLangs[j].Value ||
				(locByLangs[i].Value == locByLangs[j].Value && locByLangs[i].Key < locByLangs[j].Key)
		})
		marshalDict(w, len(locByLangs), func(i int) {
			encode(locByLangs[i].Key)
			w.WriteByte(':')
			encode(locByLangs[i].Value)
		})
	}

	if len(st.Children) != 0 {
		w.WriteString(`,"children":`)
		children := mapItems(st.Children)
		sort.Slice(children, func(i, j int) bool {
			return children[i].Value.LOC > children[j].Value.LOC ||
				(children[i].Value.LOC == children[j].Value.LOC && children[i].Key < children[j].Key)
		})
		marshalDict(w, len(children), func(i int) {
			encode(children[i].Key)
			w.WriteByte(':')
			keepErr((*SortedStat)(children[i].Value).marshalJson(w))
		})
	}
	w.WriteByte('}')
	return err
}

type KeyValue[K comparable, V any] struct {
	Key   K
	Value V
}

func mapItems[K comparable, V any](m map[K]V) []KeyValue[K, V] {
	entries := make([]KeyValue[K, V], 0, len(m))
	for key, value := range m {
		entries = append(entries, KeyValue[K, V]{key, value})
	}
	return entries
}

// marshalDict writes json dict to w.
// w must not return errors (otherwise they will be ignored).
// f must write key, ":" and value to w.
func marshalDict(w io.ByteWriter, l int, f func(i int)) {
	w.WriteByte('{')
	for i := 0; i < l; i++ {
		if i != 0 {
			w.WriteByte(',')
		}
		f(i)
	}
	w.WriteByte('}')
}

// Buffer is like bytes.Buffer, but has UnwriteByte method.
type Buffer struct {
	buf []byte
}

func (w *Buffer) Write(p []byte) (n int, err error) {
	w.buf = append(w.buf, p...)
	return len(p), nil
}

func (w *Buffer) WriteString(s string) (n int, err error) {
	w.buf = append(w.buf, s...)
	return len(s), nil
}

func (w *Buffer) WriteByte(b byte) error {
	w.buf = append(w.buf, b)
	return nil
}

func (w *Buffer) UnwriteByte() {
	w.buf = w.buf[:len(w.buf)-1]
}

func (w *Buffer) Bytes() []byte {
	return w.buf
}
