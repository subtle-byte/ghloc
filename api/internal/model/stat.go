package model

type LangName = string
type LinesNumber = int

type StatTree struct {
	LOC        LinesNumber
	LOCByLangs map[LangName]LinesNumber
	Children   map[FileName]*StatTree
}

func NewStatTreeDir() *StatTree {
	return &StatTree{
		Children: make(map[FileName]*StatTree),
	}
}

func (t StatTree) IsDir() bool {
	return t.Children != nil
}
