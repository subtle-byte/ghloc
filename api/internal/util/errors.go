package util

import (
	"fmt"
)

func WrapErr(prefix string, err *error) {
	if err == nil || *err == nil {
		return
	}
	*err = fmt.Errorf(prefix+": %w", *err)
}
