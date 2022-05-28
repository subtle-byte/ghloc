package postgres

import (
	"database/sql"
	"encoding/json"
	"ghloc/internal/github_service"
	"ghloc/internal/stat"
	"ghloc/internal/util"
	"log"
	"time"
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

func (p Postgres) SetLOCs(user, repo, branch string, locs []stat.LOCForPath) error {
	repoName := repoName(user, repo, branch)

	bytes, err := json.Marshal(locs)
	if err != nil {
		return err
	}

	start := time.Now()

	_, err = p.db.Exec("INSERT INTO repos VALUES ($1, $2, $3)", repoName, bytes, time.Now().Unix())
	if err != nil {
		return err
	}

	util.LogIOBlocking("SetLOCs", start)
	return nil
}

func (p Postgres) GetLOCs(user, repo, branch string) (locs []stat.LOCForPath, _ error) {
	repoName := repoName(user, repo, branch)

	bytes := []byte(nil)

	start := time.Now()

	err := p.db.QueryRow("SELECT locs FROM repos WHERE name = $1", repoName).Scan(&bytes)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, github_service.ErrNoData
		}
		return nil, err
	}

	util.LogIOBlocking("GetLOCs", start)

	err = json.Unmarshal(bytes, &locs)
	return locs, err
}
