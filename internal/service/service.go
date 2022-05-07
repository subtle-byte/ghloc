package service

import (
	"fmt"
	"ghloc/internal/model"
	"log"
	"time"
)

type LOCForPath struct {
	Path string
	LOC  int
}

var ErrNoData = fmt.Errorf("no data")

type LOCProvider interface {
	SetLOCs(user, repo, branch string, locs []LOCForPath) error
	GetLOCs(user, repo, branch string) ([]LOCForPath, error) // error may be ErrNoData
}

type TempStorage int

const (
	File TempStorage = iota
	Memory
)

type ContentProvider interface {
	GetContent(user, repo, branch string, tempStorage TempStorage) (*model.Content, error)
}

type Service struct {
	LOCProvider     LOCProvider // possibly nil
	ContentProvider ContentProvider
}

func (s *Service) GetStat(user, repo, branch string, filter, matcher *string, noLOCProvider bool, tempStorage TempStorage) (*model.StatTree, error) {
	if s.LOCProvider != nil {
		if !noLOCProvider {
			locs, err := s.LOCProvider.GetLOCs(user, repo, branch)
			if err == nil { // TODO?
				return buildStatTree(locs, filter, matcher), nil
			}
		} else {
			log.Println("GetStat: don't use loc provider (only in this request)")
		}
	}

	locCounter := newLOCCounter()
	contentToLOC := func(contentOpener model.Opener) (int, error) {
		rc, err := contentOpener()
		if err != nil {
			return 0, err
		}
		defer rc.Close()
		return locCounter.Count(rc)
	}

	contents, err := s.ContentProvider.GetContent(user, repo, branch, tempStorage)
	if err != nil {
		return nil, err
	}
	defer contents.Close()

	start := time.Now()

	locs := []LOCForPath(nil)
	for _, content := range contents.ByPath {
		loc, err := contentToLOC(content.ContentOpener)
		if err != nil {
			return nil, err
		}
		locs = append(locs, LOCForPath{content.Path, loc})
	}

	log.Println("LOCs counted in", time.Since(start))

	if s.LOCProvider != nil && !noLOCProvider {
		err := s.LOCProvider.SetLOCs(user, repo, branch, locs)
		if err != nil {
			log.Println("Error saving LOCs:", err)
		}
	}

	return buildStatTree(locs, filter, matcher), nil
}
