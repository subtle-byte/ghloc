package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime/pprof"
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
