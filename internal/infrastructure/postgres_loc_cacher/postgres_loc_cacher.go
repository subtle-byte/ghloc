package postgres_loc_cacher

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"time"

	"github.com/subtle-byte/ghloc/internal/service/github_stat"
	"github.com/subtle-byte/ghloc/internal/service/loc_count"
	"github.com/subtle-byte/ghloc/internal/util"
)

type Postgres struct {
	db *sql.DB
}

func NewPostgres(db *sql.DB) *Postgres {
	go func() {
		ttl := time.Hour
		for now := range time.Tick(ttl) {
			_, err := db.Exec("DELETE FROM repos WHERE cached < $1", now.Add(-ttl).Unix())
			if err != nil {
				log.Println("Error deleting old cache: ", err)
			} else {
				log.Println("Old cache is deleted")
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

	util.LogIOBlocking("SetLOCs", start)
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

	util.LogIOBlocking("GetLOCs", start)

	err = json.Unmarshal(bytes, &locs)
	return locs, err
}
