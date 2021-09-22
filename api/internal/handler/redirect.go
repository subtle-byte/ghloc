package handler

import (
	"ghloc/internal/model"
	"ghloc/internal/repository"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type RedirectHandler struct {
}

func (h *RedirectHandler) RegisterOn(router chi.Router) {
	router.Get("/{user}/{repo}", h.ServeHTTP)
}

func urlExists(url string) bool {
	resp, err := http.Head(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	switch c := resp.StatusCode; c {
	case http.StatusOK:
		return true
	case http.StatusNotFound:
		return false
	default:
		panic(c)
	}
}

func (h RedirectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	user := chi.URLParam(r, "user")
	repo := chi.URLParam(r, "repo")

	branch := ""
	if master := "master"; urlExists(repository.BuildGithubUrl(user, repo, master)) {
		branch = master
	} else if main := "main"; urlExists(repository.BuildGithubUrl(user, repo, main)) {
		branch = main
	}

	if branch == "" {
		writeResponse(w, model.BadRequest{"There is no master or main branch."})
		return
	}

	url := *r.URL
	url.Path += "/" + branch
	http.Redirect(w, r, url.String(), http.StatusTemporaryRedirect)
}
