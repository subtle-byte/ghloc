package stat

import (
	"path"
	"strings"
)

type FileName = string
type LangName = string
type LinesNumber = int

type StatTree struct {
	LOC        LinesNumber
	LOCByLangs map[LangName]LinesNumber
	Children   map[FileName]*StatTree
}

func newStatTreeDir() *StatTree {
	return &StatTree{
		Children: make(map[FileName]*StatTree),
	}
}

func (t StatTree) IsDir() bool {
	return t.Children != nil
}

func splitPath(filePath string) (dirs []string, fileName string) {
	dirsStr, fileName := path.Split(filePath)
	dirsStr = strings.TrimSuffix(dirsStr, "/")
	dirs = []string(nil)
	if dirsStr != "" {
		dirs = strings.Split(dirsStr, "/")
	}
	return
}

func (tree *StatTree) add(path string, loc int) {
	dirs, fileName := splitPath(path)

	for _, dirName := range dirs {
		child, ok := tree.Children[dirName]
		if !ok {
			child = newStatTreeDir()
			tree.Children[dirName] = child
		}
		tree = child
	}

	if fileName != "" {
		tree.Children[fileName] = &StatTree{LOC: loc}
	}
}

func getLangName(fileName string) string {
	name := path.Ext(fileName)
	if name == "" {
		return fileName
	}
	return name
}

func (dir *StatTree) countDirs() {
	dir.LOCByLangs = map[string]int{}
	for childFileName, child := range dir.Children {
		if child.IsDir() {
			child.countDirs()
			for langName, loc := range child.LOCByLangs {
				dir.LOCByLangs[langName] += loc
				dir.LOC += loc
			}
		} else {
			dir.LOCByLangs[getLangName(childFileName)] += child.LOC
			dir.LOC += child.LOC
		}
	}
}
