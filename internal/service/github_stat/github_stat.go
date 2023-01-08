package github_stat

import (
	"fmt"
	"io"
	"log"
	"time"

	"github.com/subtle-byte/ghloc/internal/service/loc_count"
)

var ErrNoData = fmt.Errorf("no data")

type LOCCacher interface {
	SetLOCs(user, repo, branch string, locs []loc_count.LOCForPath) error
	GetLOCs(user, repo, branch string) ([]loc_count.LOCForPath, error) // error may be ErrNoData
}

type TempStorage int

const (
	TempStorageFile TempStorage = iota
	TempStorageRam
)

type Opener = func() (io.ReadCloser, error)

type FileForPath struct {
	Path   string
	Opener Opener
}

type ContentProvider interface {
	GetContent(user, repo, branch string, tempStorage TempStorage) (_ []FileForPath, close func() error, _ error)
}

type Service struct {
	LOCCacher       LOCCacher // possibly nil
	ContentProvider ContentProvider
}

func (s *Service) GetStat(user, repo, branch string, filter, matcher *string, noLOCProvider bool, tempStorage TempStorage) (*loc_count.StatTree, error) {
	if s.LOCCacher != nil {
		if !noLOCProvider {
			locs, err := s.LOCCacher.GetLOCs(user, repo, branch)
			if err == nil { // TODO?
				return loc_count.BuildStatTree(locs, filter, matcher), nil
			}
		} else {
			log.Println("GetStat: don't use loc provider (only in this request)")
		}
	}

	filesForPaths, close, err := s.ContentProvider.GetContent(user, repo, branch, tempStorage)
	if err != nil {
		return nil, err
	}
	defer close()

	start := time.Now()

	locCounter := loc_count.NewFilesLOCCounter()
	for _, file := range filesForPaths {
		err := func() error {
			fileReader, err := file.Opener()
			if err != nil {
				return err
			}
			defer fileReader.Close()
			return locCounter.AddFile(file.Path, fileReader)
		}()
		if err != nil {
			return nil, err
		}
	}
	locs := locCounter.GetLOCsForPaths()

	log.Println("LOCs counted in", time.Since(start))

	if s.LOCCacher != nil && !noLOCProvider {
		err := s.LOCCacher.SetLOCs(user, repo, branch, locs)
		if err != nil {
			log.Println("Error saving LOCs:", err)
		}
	}

	return loc_count.BuildStatTree(locs, filter, matcher), nil
}
