package rest

import (
	"encoding/json"
	"errors"
	"log"
	"mime"
	"net/http"
)

type errResponse struct {
	Error string `json:"error"`
}

func WriteResponse(w http.ResponseWriter, v interface{}) {
	code := http.StatusOK

	setISE := func(err error) {
		log.Println(err)
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

	body, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		setISE(err)
	}
	w.Header().Add("Content-Type", mime.TypeByExtension(".json"))
	w.WriteHeader(code)
	_, err = w.Write(body)
	if err != nil {
		log.Println(err)
	}
}
