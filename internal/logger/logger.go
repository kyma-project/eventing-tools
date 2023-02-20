package logger

import (
	"log"
)

func LogIfError(err error) {
	if err != nil {
		log.Printf("error:[%v]", err)
	}
}

func FatalIfError(err error) {
	if err != nil {
		log.Fatalf("error:[%v]", err)
	}
}
