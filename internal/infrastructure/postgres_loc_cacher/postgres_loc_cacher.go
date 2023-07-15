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
			_, err := db.ExecContext(ctx, "DELETE FROM repos WHERE cached < $1", now.Add(-ttl).Unix())
			if err != nil {
				zerolog.Ctx(ctx).Error().Err(err).Msg("Error deleting old cache")
			} else {
				zerolog.Ctx(ctx).Info().Msg("Old cache is deleted")
			}
		}
	}()
	return &Postgres{db: db}
}

func repoName(user, repo, branch string) string {
	return user + "/" + repo + "/" + branch
}

func (p Postgres) SetLOCs(ctx context.Context, user, repo, branch string, locs []loc_count.LOCForPath) error {
	repoName := repoName(user, repo, branch)

	bytes, err := json.Marshal(locs)
	if err != nil {
		return err
	}

	start := time.Now()

	_, err = p.db.ExecContext(ctx, "INSERT INTO repos VALUES ($1, $2, $3)", repoName, bytes, time.Now().Unix())
	if err != nil {
		return err
	}

	zerolog.Ctx(ctx).Info().
		Float64("durationSec", time.Since(start).Seconds()).
		Int("sizeBytes", len(bytes)).
		Msg("Set LOC cache")
	return nil
}

func (p Postgres) GetLOCs(ctx context.Context, user, repo, branch string) (locs []loc_count.LOCForPath, _ error) {
	repoName := repoName(user, repo, branch)

	bytes := []byte(nil)

	start := time.Now()

	err := p.db.QueryRowContext(ctx, "SELECT locs FROM repos WHERE name = $1", repoName).Scan(&bytes)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, github_stat.ErrNoData
		}
		return nil, err
	}

	zerolog.Ctx(ctx).Info().
		Float64("durationSec", time.Since(start).Seconds()).
		Int("sizeBytes", len(bytes)).
		Msg("Got LOC cache")

	err = json.Unmarshal(bytes, &locs)
	return locs, err
}
