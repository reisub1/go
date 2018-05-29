package gpsparser

import (
	"log"
	"strconv"
	"strings"
	"time"
)

// Currently supports AIS140
type GPSParsed struct {
	Raw       *string // Just a copy of the raw data that was parsed
	Protocol  string  // Which protocol this message was parsed as
	Uniqid    string  // Unique identifier (Used as device ID)
	TS_Millis int64   // The timestamp from the message in unix millis
	ActualLat float32 // Latitude in conventional format (Signed)
	ActualLng float32 // Longitude in conventional format (Signed)
}

const (
	TIMEDATEFORMAT = "020106:150405"
)

// Parse function takes in a raw string and puts its GPS data in the channel
// Silently fails if it cannot parse
func Parse(raw *string, c chan *GPSParsed) {
	if *raw == "" {
		log.Printf("Empty message received")
		return
	}

	if strings.HasPrefix(*raw, "GTPL") {
		goto AIS140Parse
	} else {
		log.Printf("Invalid or unsupported protocol")
		return
	}

AIS140Parse:
	g := &GPSParsed{}
	g.Raw = raw
	g.Protocol = "AIS140"

	// This format can have multiple messages delimited by #
	messages := strings.Split(*raw, "#")

	for _, message := range messages {
		fields := strings.Split(message, ",")
		if len(fields) == 1 {
			log.Printf("Not a CSV message: %s", message)
			return
		}
		if len(fields) != 18 {
			log.Printf("Incorrect number of fields in CSV: %d", len(fields))
			return
		}
		g.Uniqid = fields[1]
		Yy_mm_dd_hh_mm_ss := strings.Join([]string{fields[3], fields[4]}, ":")
		timestamp, e := time.Parse(TIMEDATEFORMAT, Yy_mm_dd_hh_mm_ss)
		if e != nil {
			log.Printf("%s", e)
		}
		// GPSParser returns in unix seconds, but thingsboard wants it in millis
		g.TS_Millis = timestamp.Unix() * 1000

		// 5th field contains latitude as a float
		if lat, err := strconv.ParseFloat(string(fields[5]), 32); err != nil {
			log.Printf("Parsing error for latitude %s", fields[5])
			return
		} else {
			g.ActualLat = float32(lat)
		}
		// 6th field contains latitude direction information
		if strings.Compare(fields[6], "S") == 0 {
			g.ActualLat = -g.ActualLat
		}

		// 7th field contains longitude as a float
		if lng, err := strconv.ParseFloat(string(fields[7]), 32); err != nil {
			log.Printf("Parsing error for longitude %s", fields[7])
			return
		} else {
			g.ActualLng = float32(lng)
		}
		// 8th field contains lngitude direction information
		if strings.Compare(fields[8], "W") == 0 {
			g.ActualLng = -g.ActualLng
		}

		c <- g
		log.Printf("Parsed dumped")
	}
	// if len(fields) == 1 {
	// 	return fmt.Errorf("Not a CSV message: %s", message)
	// }

	// if len(fields) != 13 {
	// 	return nil, fmt.Errorf("Incorrect number of fields in CSV: %d", len(fields))
	// }
	// g.Uniqid = fields[1]

	// // Thingsboard expects time in miilis since epoch
	// _ = timestamp

	// return
}
