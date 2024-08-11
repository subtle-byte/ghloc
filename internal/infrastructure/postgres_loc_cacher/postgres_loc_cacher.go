package postgres_loc_cacher

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/rs/zerolog"
	"github.com/subtle-byte/ghloc/internal/service/github_stat"
	"github.com/subtle-byte/ghloc/internal/service/loc_count"
)

type Postgres struct {
	db *sql.DB
}

func NewPostgres(ctx context.Context, db *sql.DB) *Postgres {
	go func() {
		ttl := time.Hour
		for now := range time.Tick(ttl) {
			_, err := db.ExecContext(ctx, "DELETE FROM repos WHERE created_at < $1", now.Add(-ttl).Unix())
			if err != nil {
				zerolog.Ctx(ctx).Error().Err(err).Msg("Error deleting old cache")
			} else {
				zerolog.Ctx(ctx).Info().Msg("Old cache is deleted")
			}
		}
	}()
	return &Postgres{db: db}
}

func repoID(user, repo, branch string) string {
	return user + "/" + repo + "/" + branch
}

func (p Postgres) SetLOCs(ctx context.Context, user, repo, branch string, locs []loc_count.LOCForPath) error {
	repoID := repoID(user, repo, branch)

	locsJSON, err := json.Marshal(locs)
	if err != nil {
		return err
	}

	start := time.Now()

	_, err = p.db.ExecContext(ctx, "INSERT INTO repos VALUES ($1, $2, $3, $4)", repoID, locsJSON, false, time.Now().Unix())
	if err != nil {
		return err
	}

	zerolog.Ctx(ctx).Info().
		Float64("durationSec", time.Since(start).Seconds()).
		Int("sizeBytes", len(locsJSON)).
		Msg("Set LOC cache")
	return nil
}

func (p Postgres) SetTooLarge(ctx context.Context, user, repo, branch string) error {
	repoID := repoID(user, repo, branch)

	start := time.Now()

	_, err := p.db.ExecContext(ctx, "INSERT INTO repos VALUES ($1, $2, $3, $4)", repoID, nil, true, time.Now().Unix())
	if err != nil {
		return err
	}

	zerolog.Ctx(ctx).Info().
		Float64("durationSec", time.Since(start).Seconds()).
		Msg("Marking as too large in cache")
	return nil
}

func (p Postgres) GetLOCs(ctx context.Context, user, repo, branch string) (locs []loc_count.LOCForPath, _ error) {
	repoName := repoID(user, repo, branch)

	locBytes := []byte(nil)
	tooLarge := false

	start := time.Now()

	err := p.db.QueryRowContext(ctx, "SELECT locs, too_large FROM repos WHERE id = $1", repoName).Scan(&locBytes, &tooLarge)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, github_stat.ErrNoData
		}
		return nil, err
	}

	if tooLarge {
		return nil, github_stat.ErrRepoTooLarge
	}

	zerolog.Ctx(ctx).Info().
		Float64("durationSec", time.Since(start).Seconds()).
		Int("sizeBytes", len(locBytes)).
		Msg("Got LOC cache")

	err = json.Unmarshal(locBytes, &locs)
	return locs, err
}
