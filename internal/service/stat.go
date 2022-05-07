package service

import (
	"ghloc/internal/model"
	"path"
	"strings"
)

func splitPath(filePath string) (dirs []string, fileName string) {
	dirsStr, fileName := path.Split(filePath)
	dirsStr = strings.TrimSuffix(dirsStr, "/")
	dirs = []string(nil)
	if dirsStr != "" {
		dirs = strings.Split(dirsStr, "/")
	}
	return
}

func addInStatTree(tree *model.StatTree, locForPath LOCForPath) {
	dirs, fileName := splitPath(locForPath.Path)

	for _, dirName := range dirs {
		child, ok := tree.Children[dirName]
		if !ok {
			child = model.NewStatTreeDir()
			tree.Children[dirName] = child
		}
		tree = child
	}

	if fileName != "" {
		tree.Children[fileName] = &model.StatTree{LOC: locForPath.LOC}
	}
}

func buildStatTree(locs []LOCForPath, filter, matcher *string) *model.StatTree {
	root := model.NewStatTreeDir()

	for _, locForPath := range locs {
		if filter != nil && filtered(locForPath.Path, filter) {
			continue
		}
		if matcher != nil && !filtered(locForPath.Path, matcher) {
			continue
		}
		addInStatTree(root, locForPath)
	}

	countDirStat(root)

	return root
}

func getLangName(fileName string) string {
	name := path.Ext(fileName)
	if name == "" {
		return fileName
	}
	return name
}

func countDirStat(dir *model.StatTree) {
	dir.LOCByLangs = map[string]int{}
	for childFileName, child := range dir.Children {
		if child.IsDir() {
			countDirStat(child)
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
