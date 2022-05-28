package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime/pprof"

	// _ "net/http/pprof"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/subtle-byte/ghloc/internal/cacher/postgres"
	"github.com/subtle-byte/ghloc/internal/file_provider/github"
	"github.com/subtle-byte/ghloc/internal/github_handler"
	"github.com/subtle-byte/ghloc/internal/github_service"
)

var debugToken *string

func DebugMiddleware(next http.Handler) http.Handler {
	if debugToken == nil {
		return http.HandlerFunc(http.NotFound)
	}
	fn := func(w http.ResponseWriter, r *http.Request) {
		if r.FormValue("debug_token") == *debugToken {
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Header().Set("Content-Disposition", `attachment; filename="profile"`)
			if err := pprof.StartCPUProfile(w); err != nil {
				w.Header().Set("Content-Type", "text/plain; charset=utf-8")
				w.Header().Del("Content-Disposition")
				w.Header().Set("X-Go-Pprof", "1")
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "Could not enable CPU profiling: %s\n", err)
				return
			}
			rr := httptest.ResponseRecorder{}
			next.ServeHTTP(&rr, r)
			pprof.StopCPUProfile()
		} else {
			http.NotFound(w, r)
		}
	}
	return http.HandlerFunc(fn)
}

type MigrationLogger struct {
	Prefix string
}

func (m MigrationLogger) Printf(format string, v ...interface{}) {
	log.Print(m.Prefix, fmt.Sprintf(format, v...))
}

func (m MigrationLogger) Verbose() bool {
	return false
}

func connectDB() (_ *sql.DB, close func() error, err error) {
	dbConn := os.Getenv("DB_CONN")
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

var buildTime = "unknown" // will be replaced during building the docker image

func main() {
	log.Printf("Starting up the app (build time: %v)\n", buildTime)

	if token, ok := os.LookupEnv("DEBUG_TOKEN"); ok {
		debugToken = &token
		log.Println("Debug token is set")
	}

	github := github.Github{}
	db, closeDB, err := connectDB()
	pg := github_service.LOCProvider(nil)
	if err == nil {
		defer closeDB()
		pg = postgres.NewPostgres(db)
		log.Println("Connected to DB")
	} else {
		log.Printf("Error connecting to DB: %v", err)
		log.Println("Warning: continue without DB")
	}
	service := github_service.Service{pg, &github}

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

	getStatHandler := &github_handler.GetStatHandler{&service, debugToken}
	getStatHandler.RegisterOn(router)

	redirectHandler := &github_handler.RedirectHandler{}
	redirectHandler.RegisterOn(router)

	// router.Mount("/debug", http.DefaultServeMux)
	// router.With(DebugMiddleware).Mount("/debug", http.DefaultServeMux)
	router.With(DebugMiddleware).Route("/debug", func(r chi.Router) {
		getStatHandler.RegisterOn(r)
	})
	fmt.Println("Listening on http://localhost:8080")
	http.ListenAndServe(":8080", router)
}
