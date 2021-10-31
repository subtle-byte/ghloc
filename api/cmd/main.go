package main

import (
	"database/sql"
	"fmt"
	"ghloc/internal/handler"
	"ghloc/internal/repository"
	"ghloc/internal/service"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime/pprof"

	// _ "net/http/pprof"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/lib/pq"
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

func connectDB() (_ *sql.DB, close func() error, err error) {
	close = func() error { return nil }
	db, err := sql.Open("postgres", os.Getenv("DB_CONN"))
	if err == nil {
		close = db.Close
		err = db.Ping()
	}

	if err != nil {
		log.Printf("Error connecting to DB: %v", err)
		log.Println("Warning: continue without DB")
		close()
		return nil, nil, err
	}
	return db, close, nil
}

func main() {
	if token, ok := os.LookupEnv("DEBUG_TOKEN"); ok {
		debugToken = &token
		log.Println("Debug token is set")
	}

	github := repository.Github{}
	db, closeDB, err := connectDB()
	postgres := service.LOCProvider(nil)
	if err == nil {
		defer closeDB()
		postgres = repository.NewPostgres(db)
	}
	service := service.Service{postgres, &github}

	router := chi.NewRouter()
	router.Use(middleware.RealIP)
	middleware.RequestIDHeader = ""
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Compress(5))
	router.Use(middleware.Recoverer)

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "<html><body><a href='https://github.com/subtle-byte/ghloc'>Docs</a></body><html>")
	})

	getStatHandler := &handler.GetStatHandler{&service, debugToken}
	getStatHandler.RegisterOn(router)

	redirectHandler := &handler.RedirectHandler{}
	redirectHandler.RegisterOn(router)

	// router.Mount("/debug", http.DefaultServeMux)
	// router.With(DebugMiddleware).Mount("/debug", http.DefaultServeMux)
	router.With(DebugMiddleware).Route("/debug", func(r chi.Router) {
		getStatHandler.RegisterOn(r)
	})

	http.ListenAndServe(":8080", router)
}
