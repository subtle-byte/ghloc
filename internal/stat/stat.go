package stat

func BuildStatTree(locsForPaths []LOCForPath, filter, matcher *string) *StatTree {
	root := newStatTreeDir()

	for _, locForPath := range locsForPaths {
		if filter != nil && filtered(locForPath.Path, filter) {
			continue
		}
		if matcher != nil && !filtered(locForPath.Path, matcher) {
			continue
		}
		root.add(locForPath.Path, locForPath.LOC)
	}

	root.countDirs()

	return root
}
