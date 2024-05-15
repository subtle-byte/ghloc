package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime/pprof"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
	"github.com/subtle-byte/ghloc/internal/util"
)

func NewDebugMiddleware(debugToken string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if debugToken == "" {
			return http.HandlerFunc(http.NotFound)
		}
		fn := func(w http.ResponseWriter, r *http.Request) {
			if r.FormValue("debug_token") == debugToken {
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
}

func NewLoggerMiddleware(logger zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			reqID := middleware.GetReqID(ctx)
			logger := logger.With().Str("requestId", reqID).Logger()
			ctx = logger.WithContext(ctx)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

func NewRequestLoggerMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			logger := *zerolog.Ctx(r.Context())
			logger = logger.With().
				Str("method", r.Method).
				Bool("tls", r.TLS != nil).
				Str("host", r.Host).
				Str("url", r.RequestURI).
				Str("protocol", r.Proto).
				Str("from", r.RemoteAddr).
				Str("origin", r.Header.Get("Origin")).
				Logger()
			logger.Info().Msg("New request")
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			defer func() {
				logger.Info().
					Float64("durationSec", time.Since(start).Seconds()).
					Int("status", ww.Status()).
					Int("responseBytes", ww.BytesWritten()).
					Msg("Request finished")
			}()
			next.ServeHTTP(ww, r)
		}
		return http.HandlerFunc(fn)
	}
}

func NewRecoverMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					stack := util.GetStack(1)
					zerolog.Ctx(r.Context()).Error().
						Any("stack", stack).
						Any("panicValue", err).
						Msg("Panic recovered")
					w.WriteHeader(http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}
