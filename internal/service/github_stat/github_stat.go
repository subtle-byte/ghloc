package github_stat

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/rs/zerolog"
	"github.com/subtle-byte/ghloc/internal/service/loc_count"
	"github.com/subtle-byte/ghloc/internal/util"
)

var ErrNoData = fmt.Errorf("no data")

type LOCCacher interface {
	SetLOCs(ctx context.Context, user, repo, branch string, locs []loc_count.LOCForPath) error
	GetLOCs(ctx context.Context, user, repo, branch string) ([]loc_count.LOCForPath, error) // error may be ErrNoData
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
	GetContent(ctx context.Context, user, repo, branch string, tempStorage TempStorage) (_ []FileForPath, close func() error, _ error)
}

type Service struct {
	LOCCacher       LOCCacher // possibly nil
	ContentProvider ContentProvider
	sem             chan struct{} // semaphore for limiting number of concurrent work
}

func New(locCacher LOCCacher, contentProvider ContentProvider, maxParallelWork int) *Service {
	return &Service{
		LOCCacher:       locCacher,
		ContentProvider: contentProvider,
		sem:             make(chan struct{}, maxParallelWork),
	}
}

func (s *Service) GetStat(ctx context.Context, user, repo, branch string, filter, matcher *string, noLOCProvider bool, tempStorage TempStorage) (*loc_count.StatTree, error) {
	select {
	case s.sem <- struct{}{}:
		defer func() { <-s.sem }()
	case <-ctx.Done():
		return nil, fmt.Errorf("wait in queue: %w", ctx.Err())
	}

	var locs *[]loc_count.LOCForPath
	if s.LOCCacher != nil {
		if !noLOCProvider {
			cacheLocs, err := s.LOCCacher.GetLOCs(ctx, user, repo, branch)
			if err != nil {
				zerolog.Ctx(ctx).Warn().Err(err).Msg("Failed to get LOCs from cache, will proceed without it")
			} else {
				locs = &cacheLocs
			}
		} else {
			zerolog.Ctx(ctx).Info().Msg("Proceeding without LOCs cache - as requested")
		}
	}

	if locs == nil {
		filesForPaths, close, err := s.ContentProvider.GetContent(ctx, user, repo, branch, tempStorage)
		if err != nil {
			return nil, fmt.Errorf("get repo content: %w", err)
		}
		defer close()

		locCountingStart := time.Now()

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
		locs = util.Pointer(locCounter.GetLOCsForPaths())
		zerolog.Ctx(ctx).Info().
			Float64("durationSec", time.Since(locCountingStart).Seconds()).
			Msg("LOCs counted")

		if s.LOCCacher != nil && !noLOCProvider {
			err := s.LOCCacher.SetLOCs(ctx, user, repo, branch, *locs)
			if err != nil {
				zerolog.Ctx(ctx).Error().Err(err).Msg("Error saving LOCs to the cache")
			}
		}
	}

	treeBuildingStart := time.Now()
	statTree := loc_count.BuildStatTree(*locs, filter, matcher)
	zerolog.Ctx(ctx).Info().
		Float64("durationSec", time.Since(treeBuildingStart).Seconds()).
		Msg("Statistics tree built")

	return statTree, nil
}
