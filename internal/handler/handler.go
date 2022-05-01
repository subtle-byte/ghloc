package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"ghloc/internal/model"
	"ghloc/internal/service"

	"github.com/go-chi/chi/v5"
)

type Service interface {
	GetStat(user, repo, branch string, filter []string, noLOCProvider bool, tempStorage service.TempStorage) (*model.StatTree, error)
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
	tempStorage := service.File
	if h.DebugToken != nil {
		debugTokenInRequest := r.FormValue("debug_token")
		if debugTokenInRequest == *h.DebugToken {
			if r.Form["no_cache"] != nil {
				noLOCProvider = true
			}
			if r.Form["mem_for_temp"] != nil {
				tempStorage = service.Memory
			}
		} else if debugTokenInRequest != "" {
			writeResponse(w, model.BadRequest{"Invalid debug token"})
			return
		}
	}

	filter := r.Form["filter"]
	if len(filter) == 1 {
		filter = strings.Split(filter[0], ",")
	}

	stat, err := h.Service.GetStat(user, repo, branch, filter, noLOCProvider, tempStorage)
	if err != nil {
		writeResponse(w, err)
		return
	}
	writeResponse(w, (*response)(stat))
}

type response model.StatTree

func (stat *response) MarshalJSON() ([]byte, error) {
	if stat.Children == nil {
		return json.Marshal(stat.LOC)
	} else {
		resp := struct {
			LOC        int       `json:"loc"`
			LOCByLangs SortedMap `json:"locByLangs,omitempty"`
			Children   SortedMap `json:"children,omitempty"`
		}{
			stat.LOC,
			SortedMap{
				stat.LOCByLangs,
				nil,
				func(loc1, loc2 interface{}) bool {
					return loc1.(model.LinesNumber) > loc2.(model.LinesNumber)
				},
			},
			SortedMap{
				stat.Children,
				func(value interface{}) interface{} {
					return (*response)(value.(*model.StatTree))
				},
				func(stat1, stat2 interface{}) bool {
					return stat1.(*response).LOC > stat2.(*response).LOC
				},
			},
		}
		return json.Marshal(resp)
	}
}
