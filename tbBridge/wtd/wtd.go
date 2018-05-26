package wtd

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type WTD struct {
	csvParsed []string
	// Identifierstring "ZJ" //`field0`
	Uniqid string //`field1`
	// Version   string    //`field2`
	// Hh_mm_ss  string    //`field3`
	// GpsValid  string    //`field4`
	// Lat       string    //`field5`
	// Latdir    string    //`field6`
	// Lng       string    //`field7`
	// Lngdir    string    //`field8`
	// Speed     string    //`field9`
	// Angle     string    //`field10`
	// Yy_mm_dd  string    //`field11`
	// Status    string    //`field12`
	TS_Millis int64   //`Virtual field to pass outside`
	ActualLat float32 //`Virtual field to pass outside`
	ActualLng float32 //`Virtual field to pass outside`
}

const (
	TIMEDATEFORMAT = "150405:020106"
)

func Parse(message string) (parsed *WTD, e error) {
	e = nil

	if strings.Contains(message, "*ZJ#") {
		return nil, fmt.Errorf("Empty message")
	}

	parsed = &WTD{}
	parsed.csvParsed = strings.Split(message, ",")
	if len(parsed.csvParsed) == 1 {
		return nil, fmt.Errorf("Not a CSV message: %s", message)
	}

	if len(parsed.csvParsed) != 13 {
		return nil, fmt.Errorf("Incorrect number of fields in CSV: %d", len(parsed.csvParsed))
	}
	parsed.Uniqid = parsed.csvParsed[1]

	Hh_mm_ss_Yy_mm_dd := strings.Join([]string{parsed.csvParsed[3], parsed.csvParsed[11]}, ":")
	timestamp, e := time.ParseInLocation(TIMEDATEFORMAT, Hh_mm_ss_Yy_mm_dd, time.Local)
	// WTD returns in unix seconds, we want in millis
	parsed.TS_Millis = timestamp.Unix() * 1000
	fmt.Println(time.Unix(timestamp.Unix(), 0))

	// 6th field contains latitude as a float
	if lat, err := strconv.ParseFloat(string(parsed.csvParsed[5]), 32); err != nil {
		return nil, fmt.Errorf("Parsing error for latitude %s", parsed.csvParsed[5])
	} else {
		parsed.ActualLat = float32(lat) / 100
	}
	// 7th field contains latitude direction information
	if strings.Compare(parsed.csvParsed[6], "S") == 0 {
		parsed.ActualLat = -1 * parsed.ActualLat
	}

	if lng, err := strconv.ParseFloat(string(parsed.csvParsed[7]), 32); err != nil {
		return nil, fmt.Errorf("Parsing error for longitude %s", parsed.csvParsed[7])
	} else {
		parsed.ActualLng = float32(lng) / 100
	}
	// 7th field contains lngitude direction information
	if strings.Compare(parsed.csvParsed[8], "W") == 0 {
		parsed.ActualLng = -1 * parsed.ActualLng
	}

	// Thingsboard expects time in miilis since epoch
	_ = timestamp

	return
}
