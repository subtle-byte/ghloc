package rest

import "fmt"

var NotFound = fmt.Errorf("not found")

type BadRequest struct {
	Msg string
}

func (e BadRequest) Error() string {
	return e.Msg
}
