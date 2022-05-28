package rest

import (
	"encoding/json"

	"github.com/subtle-byte/ghloc/internal/stat"
)

type Stat stat.StatTree

func (st *Stat) MarshalJSON() ([]byte, error) {
	if st.Children == nil {
		return json.Marshal(st.LOC)
	} else {
		resp := struct {
			LOC        int       `json:"loc"`
			LOCByLangs SortedMap `json:"locByLangs,omitempty"`
			Children   SortedMap `json:"children,omitempty"`
		}{
			st.LOC,
			SortedMap{
				st.LOCByLangs,
				nil,
				func(loc1, loc2 interface{}) bool {
					return loc1.(stat.LinesNumber) > loc2.(stat.LinesNumber)
				},
			},
			SortedMap{
				st.Children,
				func(value interface{}) interface{} {
					return (*Stat)(value.(*stat.StatTree))
				},
				func(stat1, stat2 interface{}) bool {
					return stat1.(*Stat).LOC > stat2.(*Stat).LOC
				},
			},
		}
		return json.Marshal(resp)
	}
}
