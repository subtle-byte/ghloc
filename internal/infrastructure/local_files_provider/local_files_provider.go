package local_files_provider

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/subtle-byte/ghloc/internal/service/github_stat"
)

func GetFilesInDir(path string) ([]github_stat.FileForPath, error) {
	ffp := []github_stat.FileForPath(nil)
	err := filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == ".git" || path == ".vscode" || path == ".idea" {
			return fs.SkipDir
		}
		if !d.Type().IsRegular() {
			return nil
		}
		ffp = append(ffp, github_stat.FileForPath{
			Path: path,
			Opener: func() (io.ReadCloser, error) {
				return os.Open(path)
			},
		})
		return nil
	})
	return ffp, err
}
