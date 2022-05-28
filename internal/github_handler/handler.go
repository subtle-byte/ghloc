package github_handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/subtle-byte/ghloc/internal/github_service"
	"github.com/subtle-byte/ghloc/internal/rest"
	"github.com/subtle-byte/ghloc/internal/stat"
)

type Service interface {
	GetStat(user, repo, branch string, filter, matcher *string, noLOCProvider bool, tempStorage github_service.TempStorage) (*stat.StatTree, error)
}

type GetStatHandler struct {
	Service    Service
	DebugToken *string
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
	tempStorage := github_service.TempStorageFile
	if h.DebugToken != nil {
		debugTokenInRequest := r.FormValue("debug_token")
		if debugTokenInRequest == *h.DebugToken {
			if r.Form["no_cache"] != nil {
				noLOCProvider = true
			}
			if r.Form["mem_for_temp"] != nil {
				tempStorage = github_service.TempStorageRam
			}
		} else if debugTokenInRequest != "" {
			rest.WriteResponse(w, rest.BadRequest{"Invalid debug token"})
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

	stat, err := h.Service.GetStat(user, repo, branch, filter, matcher, noLOCProvider, tempStorage)
	if err != nil {
		rest.WriteResponse(w, err)
		return
	}
	rest.WriteResponse(w, (*rest.Stat)(stat))
}
