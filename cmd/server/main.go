package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	// _ "net/http/pprof"

	"github.com/caarlos0/env/v9"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
	"github.com/subtle-byte/ghloc/internal/infrastructure/github_files_provider"
	"github.com/subtle-byte/ghloc/internal/infrastructure/postgres_loc_cacher"
	"github.com/subtle-byte/ghloc/internal/server/github_handler"
	github_stat_service "github.com/subtle-byte/ghloc/internal/service/github_stat"
)

type Config struct {
	JSONLogs          bool   `env:"JSON_LOGS" envDefault:"true"`
	DebugToken        string `env:"DEBUG_TOKEN"`
	MaxRepoSizeMB     int    `env:"MAX_REPO_SIZE_MB,notEmpty"`
	MaxConcurrentWork int    `env:"MAX_CONCURRENT_WORK,notEmpty"`
	DbConnStr         string `env:"DB_CONN"`
}

var buildTime = "unknown" // will be replaced during building the docker image

func main() {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerolog.TimestampFunc = func() time.Time {
		return time.Now().Round(time.Microsecond).UTC()
	}
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	log.SetFlags(0)
	log.SetOutput(logger)

	logger.Info().Str("buildTime", buildTime).Msg("Starting up the app")

	ctx := context.Background()
	ctx = logger.WithContext(ctx)

	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		logger.Fatal().Err(err).Msg("Error parsing config")
	}
	if !cfg.JSONLogs {
		out := &zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: zerolog.TimeFieldFormat}
		logger = logger.Output(out)
	}
	logger.Info().Msgf("Debug token is set: %v", cfg.DebugToken != "")

	github := github_files_provider.New(cfg.MaxRepoSizeMB)
	db, closeDB, err := connectAndMigrateDB(cfg.DbConnStr)
	pg := github_stat_service.LOCCacher(nil)
	if err == nil {
		defer closeDB()
		pg = postgres_loc_cacher.NewPostgres(ctx, db)
		logger.Info().Msg("Connected to DB")
	} else {
		logger.Info().Err(err).Msg("Error connecting to DB")
		logger.Warn().Msg("Warning: continue without DB")
	}
	service := github_stat_service.New(pg, github, cfg.MaxConcurrentWork)

	router := chi.NewRouter()
	router.Use(middleware.RealIP)
	middleware.RequestIDHeader = ""
	router.Use(middleware.RequestID)
	router.Use(NewLoggerMiddleware(logger))
	router.Use(NewRequestLoggerMiddleware())
	router.Use(middleware.Compress(5))
	router.Use(NewRecoverMiddleware())
	router.Use(cors.AllowAll().Handler)

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "<html><body><a href='https://github.com/subtle-byte/ghloc'>Docs</a></body><html>")
	})

	getStatHandler := &github_handler.GetStatHandler{Service: service, DebugToken: cfg.DebugToken}
	getStatHandler.RegisterOn(router)

	redirectHandler := &github_handler.RedirectHandler{}
	redirectHandler.RegisterOn(router)

	// router.Mount("/debug", http.DefaultServeMux)
	// router.With(DebugMiddleware).Mount("/debug", http.DefaultServeMux)
	router.With(NewDebugMiddleware(cfg.DebugToken)).Route("/debug", func(r chi.Router) {
		getStatHandler.RegisterOn(r)
	})
	logger.Info().Msg("Listening on http://localhost:8080")
	http.ListenAndServe(":8080", router)
}
