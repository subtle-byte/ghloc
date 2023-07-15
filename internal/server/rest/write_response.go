package rest

import (
	"encoding/json"
	"errors"
	"mime"
	"net/http"

	"github.com/rs/zerolog"
)

type errResponse struct {
	Error string `json:"error"`
}

func WriteResponse(w http.ResponseWriter, r *http.Request, v interface{}, pretty bool) {
	code := http.StatusOK

	setISE := func(err error) {
		zerolog.Ctx(r.Context()).Error().Err(err).Msg("Internal server error")
		code = http.StatusInternalServerError
		v = errResponse{"Internal server error"}
	}

	if err, ok := v.(error); ok {
		if badRequest := (BadRequest{}); errors.As(err, &badRequest) {
			code = http.StatusBadRequest
			v = errResponse{badRequest.Msg}
		} else if errors.Is(err, NotFound) {
			code = http.StatusNotFound
			v = errResponse{"Not found"}
		} else {
			setISE(err)
		}
	}

	var body []byte
	var err error
	if pretty {
		body, err = json.MarshalIndent(v, "", "  ")
	} else {
		body, err = json.Marshal(v)
	}
	if err != nil {
		setISE(err)
	}
	w.Header().Add("Content-Type", mime.TypeByExtension(".json"))
	w.WriteHeader(code)
	_, err = w.Write(body)
	if err != nil {
		zerolog.Ctx(r.Context()).Error().Err(err).Msg("Error writing http response")
	}
}
