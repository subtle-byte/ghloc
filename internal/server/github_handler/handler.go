package github_handler

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/subtle-byte/ghloc/internal/server/rest"
	"github.com/subtle-byte/ghloc/internal/service/github_stat"
	"github.com/subtle-byte/ghloc/internal/service/loc_count"
)

type Service interface {
	GetStat(ctx context.Context, user, repo, branch string, filter, matcher *string, noLOCProvider bool, tempStorage github_stat.TempStorage) (*loc_count.StatTree, error)
}

type GetStatHandler struct {
	Service    Service
	DebugToken string
}

func (h *GetStatHandler) RegisterOn(router chi.Router) {
	router.Get("/{user}/{repo}/{branch}", h.ServeHTTP)
}

func (h GetStatHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	user := chi.URLParam(r, "user")
	repo := chi.URLParam(r, "repo")
	branch := chi.URLParam(r, "branch")

	r.ParseForm()

	noLOCProvider := false
	tempStorage := github_stat.TempStorageFile
	if h.DebugToken != "" {
		debugTokenInRequest := r.FormValue("debug_token")
		if debugTokenInRequest == h.DebugToken {
			if r.Form["no_cache"] != nil {
				noLOCProvider = true
			}
			if r.Form["mem_for_temp"] != nil {
				tempStorage = github_stat.TempStorageRam
			}
		} else if debugTokenInRequest != "" {
			rest.WriteResponse(w, r, rest.BadRequest{"Invalid debug token"}, true)
			return
		}
	}

	var filter *string
	if filters := r.Form["filter"]; len(filters) >= 1 {
		filter = &filters[0]
	}

	var matcher *string
	if matchers := r.Form["match"]; len(matchers) >= 1 {
		matcher = &matchers[0]
	}

	stat, err := h.Service.GetStat(r.Context(), user, repo, branch, filter, matcher, noLOCProvider, tempStorage)
	if err != nil {
		rest.WriteResponse(w, r, err, true)
		return
	}
	w.Header().Add("Cache-Control", "public, max-age=300")
	rest.WriteResponse(w, r, (*rest.SortedStat)(stat), r.FormValue("pretty") != "false")
}
