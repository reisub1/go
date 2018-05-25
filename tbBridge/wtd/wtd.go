package wtd

import (
	"errors"
	"log"
	"strings"
)

type WTD struct {
	Uniqid   string
	Version  [2]byte
	Hh_mm_ss [6]byte
	GpsValid byte
	Lat      float32
	Latdir   byte
	Lng      float32
	Lngdir   byte
	Speed    float32
	Angle    float32
	Yy_mm_dd [6]byte
	Status   int64
}

func Parse(message string) (parsed *WTD, e error) {
	var errorString = "Invalid Message"
	e = errors.New(errorString)
	if strings.Contains(message, "*ZJ#") {
		return
	}

	tokens := strings.Split(message, ",")
	parsed = &WTD{}
	if len(tokens) == 1 {
		log.Printf("Invalid message(%d) received: %s", len(message), message)
		return
	}

	if len(tokens[1]) != 10 {
		return
	}

	parsed.Uniqid = tokens[1]

	e = nil
	return
}
