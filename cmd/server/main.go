package main

import (
	"fmt"
	"log"
	"net/http"

	// _ "net/http/pprof"

	"github.com/caarlos0/env/v9"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/subtle-byte/ghloc/internal/infrastructure/github_files_provider"
	"github.com/subtle-byte/ghloc/internal/infrastructure/postgres_loc_cacher"
	"github.com/subtle-byte/ghloc/internal/server/github_handler"
	github_stat_service "github.com/subtle-byte/ghloc/internal/service/github_stat"
)

type Config struct {
	DebugToken        string `env:"DEBUG_TOKEN"`
	MaxRepoSizeMB     int    `env:"MAX_REPO_SIZE_MB,notEmpty"`
	MaxConcurrentWork int    `env:"MAX_CONCURRENT_WORK,notEmpty"`
	DbConnStr         string `env:"DB_CONN"`
}

var buildTime = "unknown" // will be replaced during building the docker image

func main() {
	log.Printf("Starting up the app (build time: %v)\n", buildTime)

	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		log.Fatalf("Parsing config: %v", err)
	}
	log.Printf("Debug token is set: %v", cfg.DebugToken != "")

	github := github_files_provider.New(cfg.MaxRepoSizeMB)
	db, closeDB, err := connectAndMigrateDB(cfg.DbConnStr)
	pg := github_stat_service.LOCCacher(nil)
	if err == nil {
		defer closeDB()
		pg = postgres_loc_cacher.NewPostgres(db)
		log.Println("Connected to DB")
	} else {
		log.Printf("Error connecting to DB: %v", err)
		log.Println("Warning: continue without DB")
	}
	service := github_stat_service.New(pg, github, cfg.MaxConcurrentWork)

	router := chi.NewRouter()
	router.Use(middleware.RealIP)
	middleware.RequestIDHeader = ""
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Compress(5))
	router.Use(middleware.Recoverer)
	router.Use(cors.AllowAll().Handler)

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "<html><body><a href='https://github.com/subtle-byte/ghloc'>Docs</a></body><html>")
	})

	getStatHandler := &github_handler.GetStatHandler{service, cfg.DebugToken}
	getStatHandler.RegisterOn(router)

	redirectHandler := &github_handler.RedirectHandler{}
	redirectHandler.RegisterOn(router)

	// router.Mount("/debug", http.DefaultServeMux)
	// router.With(DebugMiddleware).Mount("/debug", http.DefaultServeMux)
	router.With(NewDebugMiddleware(cfg.DebugToken)).Route("/debug", func(r chi.Router) {
		getStatHandler.RegisterOn(r)
	})
	fmt.Println("Listening on http://localhost:8080")
	http.ListenAndServe(":8080", router)
}
