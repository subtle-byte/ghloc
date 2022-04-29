package handler

import (
	"encoding/json"
	"fmt"
	"ghloc/internal/model"
	"ghloc/internal/util"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type RedirectHandler struct {
}

func (h *RedirectHandler) RegisterOn(router chi.Router) {
	router.Get("/{user}/{repo}", h.ServeHTTP)
}

func getDefaultBranch(user, repo string) (_ string, err error) {
	defer util.WrapErr("get default branch", &err)

	resp, err := http.Get(fmt.Sprintf("https://api.github.com/repos/%v/%v", user, repo))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusNotFound {
		return "", model.NotFound
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status %s (body: %s)", resp.Status, string(body))
	}

	repoInfo := struct {
		DefaultBranch string `json:"default_branch"`
	}{}
	if err = json.Unmarshal(body, &repoInfo); err != nil {
		return "", err
	}
	if repoInfo.DefaultBranch == "" {
		return "", fmt.Errorf("empty branch (body: %s)", string(body))
	}

	return repoInfo.DefaultBranch, nil
}

func (h RedirectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	user := chi.URLParam(r, "user")
	repo := chi.URLParam(r, "repo")

	branch, err := getDefaultBranch(user, repo)
	if err != nil {
		writeResponse(w, err)
		return
	}

	url := *r.URL
	url.Path += "/" + branch
	http.Redirect(w, r, url.String(), http.StatusTemporaryRedirect)
}
