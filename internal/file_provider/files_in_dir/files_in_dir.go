package files_in_dir

import (
	"ghloc/internal/file_provider"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

func GetFilesInDir(path string) ([]file_provider.FileForPath, error) {
	ffp := []file_provider.FileForPath(nil)
	err := filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == ".git" || path == ".vscode" || path == ".idea" {
			return fs.SkipDir
		}
		if d.IsDir() {
			return nil
		}
		ffp = append(ffp, file_provider.FileForPath{
			Path: path,
			Opener: func() (io.ReadCloser, error) {
				return os.Open(path)
			},
		})
		return nil
	})
	return ffp, err
}
