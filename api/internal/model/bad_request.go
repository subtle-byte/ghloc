package model

type BadRequest struct {
	Msg string
}

func (e BadRequest) Error() string {
	return e.Msg
}
