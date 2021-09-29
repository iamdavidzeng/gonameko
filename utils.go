package gonameko

import (
	"log"
	"strings"
)

func FailOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func ToMethodName(str string) string {
	slc := strings.Split(str, "_")
	for i := range slc {
		slc[i] = strings.Title(slc[i])
	}
	return strings.Join(slc, "")
}
