package rest

import (
	"encoding/json"

	"github.com/subtle-byte/ghloc/internal/service/loc_count"
)

type Stat loc_count.StatTree

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
					return loc1.(loc_count.LinesNumber) > loc2.(loc_count.LinesNumber)
				},
			},
			SortedMap{
				st.Children,
				func(value interface{}) interface{} {
					return (*Stat)(value.(*loc_count.StatTree))
				},
				func(stat1, stat2 interface{}) bool {
					return stat1.(*Stat).LOC > stat2.(*Stat).LOC
				},
			},
		}
		return json.Marshal(resp)
	}
}
