package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
)

type MigrationLogger struct {
	Prefix string
}

func (m MigrationLogger) Printf(format string, v ...interface{}) {
	log.Print(m.Prefix, fmt.Sprintf(format, v...))
}

func (m MigrationLogger) Verbose() bool {
	return false
}

func connectAndMigrateDB(dbConn string) (_ *sql.DB, close func() error, err error) {
	if dbConn == "" {
		return nil, nil, fmt.Errorf("env var DB_CONN is not provided")
	}

	m, err := migrate.New("file://migrations", dbConn)
	if err != nil {
		return nil, nil, fmt.Errorf("create migrator: %w", err)
	}
	m.Log = MigrationLogger{Prefix: "migration: "}
	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return nil, nil, fmt.Errorf("migrate up: %w", err)
	}

	close = func() error { return nil }
	db, err := sql.Open("postgres", dbConn)
	if err == nil {
		close = db.Close
		err = db.Ping()
	}

	if err != nil {
		close()
		return nil, nil, fmt.Errorf("connect to db: %w", err)
	}
	return db, close, nil
}
