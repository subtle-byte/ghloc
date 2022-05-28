package github_service

import (
	"fmt"
	"log"
	"time"

	"github.com/subtle-byte/ghloc/internal/file_provider"
	"github.com/subtle-byte/ghloc/internal/stat"
)

var ErrNoData = fmt.Errorf("no data")

type LOCProvider interface {
	SetLOCs(user, repo, branch string, locs []stat.LOCForPath) error
	GetLOCs(user, repo, branch string) ([]stat.LOCForPath, error) // error may be ErrNoData
}

type TempStorage int

const (
	TempStorageFile TempStorage = iota
	TempStorageRam
)

type ContentProvider interface {
	GetContent(user, repo, branch string, tempStorage TempStorage) (_ []file_provider.FileForPath, close func() error, _ error)
}

type Service struct {
	LOCProvider     LOCProvider // possibly nil
	ContentProvider ContentProvider
}

func (s *Service) GetStat(user, repo, branch string, filter, matcher *string, noLOCProvider bool, tempStorage TempStorage) (*stat.StatTree, error) {
	if s.LOCProvider != nil {
		if !noLOCProvider {
			locs, err := s.LOCProvider.GetLOCs(user, repo, branch)
			if err == nil { // TODO?
				return stat.BuildStatTree(locs, filter, matcher), nil
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

	locCounter := stat.NewLOCCounter()
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

	if s.LOCProvider != nil && !noLOCProvider {
		err := s.LOCProvider.SetLOCs(user, repo, branch, locs)
		if err != nil {
			log.Println("Error saving LOCs:", err)
		}
	}

	return stat.BuildStatTree(locs, filter, matcher), nil
}
