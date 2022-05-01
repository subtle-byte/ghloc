package repository

import (
	"fmt"
	"log"
	"time"
)

func logIOBlocking(name string, start time.Time, meta ...interface{}) {
	metaStr := ""
	if meta != nil {
		metaStr += "("
		for i, m := range meta {
			if i != 0 {
				metaStr += ", "
			}
			metaStr += fmt.Sprint(m)
		}
		metaStr += ")"
	}
	log.Println(name+": waited for IO for", time.Since(start), metaStr)
}
