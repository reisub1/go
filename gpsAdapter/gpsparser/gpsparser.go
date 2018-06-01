package gpsparser

import (
	"bytes"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

// Currently supports AIS140
type GPSParsed struct {
	Raw            *string // Just a copy of the raw data that was parsed
	Protocol       string  // Which protocol this message was parsed as
	PacketType     string  // Type of packet (Status? Alert? OverSpeed?)
	Uniqid         string  // Unique identifier (Used as device ID)
	TS_Millis      int64   // The timestamp from the message in unix millis
	ActualLat      float64 // Latitude in conventional format (Signed)
	ActualLng      float64 // Longitude in conventional format (Signed)
	Speed          int
	OdoMeter       int
	Direction      int
	NoOfSatellites int
	StatusBox      bool    //true = Box Open, false = Box Closed
	GSMSignal      int     // Signal Strength
	StatusBattery  bool    // true = Battery Connected, false = Battery Disconnected
	BatteryLow     bool    // true = Battery Low, false = Battery Normal
	StatusIgnition bool    // true = Ignition on, false = Ignition off
	Voltage        float64 // Analog voltage
}

const (
	TIMEDATEFORMAT = "020106:150405"
	DEBUGGING      = false
)

// Parse function takes in a raw string and puts its GPS data in the channel
// Silently fails if it cannot parse
func Parse(raw *string, c chan *string) {
	if *raw == "" {
		if DEBUGGING {
			log.Printf("Empty message received")
		}
		return
	}

	if strings.HasPrefix(*raw, "GTPL") {
		goto AIS140Parse
	} else {
		if DEBUGGING {
			log.Printf("Invalid or unsupported protocol")
		}
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
			if DEBUGGING {
				log.Printf("Not a CSV message: %s", message)
			}
			continue
		}
		switch len(fields) {
		case 10:
			break
		case 18:
			break
		default:
			if DEBUGGING {
				log.Printf("Invalid number of fields in CSV: %s", message)
			}
			continue

		}

		// 0th field has GTPL $1, GTPL $2 etc for different types of packets
		g.PacketType = strings.Split(fields[0], " ")[1]

		// For any packet type we have the following fields Uniqid, TimeDate, Lat, Long
		g.Uniqid = fields[1]
		Yy_mm_dd_hh_mm_ss := strings.Join([]string{fields[3], fields[4]}, ":")
		timestamp, e := time.Parse(TIMEDATEFORMAT, Yy_mm_dd_hh_mm_ss)
		if e != nil {
			if DEBUGGING {
				log.Printf("%s", e)
			}
		}
		// GPSParser returns in unix seconds, but thingsboard wants it in millis
		g.TS_Millis = timestamp.Unix() * 1000

		// 5th field contains latitude as a float
		if lat, err := strconv.ParseFloat(fields[5], 64); err != nil {
			if DEBUGGING {
				log.Printf("Parsing error for latitude %s", fields[5])
			}
			continue
		} else {
			g.ActualLat = float64(lat)
		}
		// 6th field contains latitude direction information
		if fields[6] == "S" {
			g.ActualLat = -g.ActualLat
		}

		// 7th field contains longitude as a float
		if lng, err := strconv.ParseFloat(fields[7], 64); err != nil {
			if DEBUGGING {
				log.Printf("Parsing error for longitude %s", fields[7])
			}
			continue
		} else {
			g.ActualLng = float64(lng)
		}
		// 8th field contains lngitude direction information
		if fields[8] == "W" {
			g.ActualLng = -g.ActualLng
		}

		var jsonBuffer bytes.Buffer
		jsonBuffer.WriteString("{") // Start the Json Object
		// Add whatever we've parsed so far into the JSON Object
		jsonBuffer.WriteString(fmt.Sprintf(`"%s":[{"ts":%d,"values":{"lat":%f,"lng":%f`, g.Uniqid, g.TS_Millis, g.ActualLat, g.ActualLng))
		//Note that no comma has been inserted at the end

		// Now each packetType has its own specific parameters and syntax
		switch g.PacketType {
		// Status packet ($1)
		case "$1":
			if speed, err := strconv.ParseInt(fields[9], 10, 64); err != nil {
				if DEBUGGING {
					log.Printf("Parsing error for speed %s", fields[9])
				}
				continue
			} else {
				g.Speed = int(speed)

				// Add the parsed Speed to the json
				jsonBuffer.WriteString(fmt.Sprintf(`,"speed":%d`, g.Speed))
			}

			if boxOpen, err := strconv.ParseBool(fields[13]); err != nil {
				if DEBUGGING {
					log.Printf("Parsing error for boxOpen %s", fields[13])
				}
				continue
			} else {
				g.StatusBox = boxOpen

				// Add the box Status to json
				// Adds true or false (Json booleans)
				jsonBuffer.WriteString(fmt.Sprintf(`,"box":%v`, g.StatusBox))
			}

			if batConnected, err := strconv.ParseBool(fields[15]); err != nil {
				if DEBUGGING {
					log.Printf("Parsing error for batConnected %s", fields[15])
				}
				continue
			} else {
				g.StatusBattery = batConnected

				// Add the battery Status to json
				// Adds true or false (Json booleans)
				jsonBuffer.WriteString(fmt.Sprintf(`,"bat":%v`, g.StatusBattery))
			}

			if ignition, err := strconv.ParseBool(fields[16]); err != nil {
				if DEBUGGING {
					log.Printf("Parsing error for ignition %s", fields[16])
				}
				continue
			} else {
				g.StatusIgnition = ignition

				// Add the ignition Status to json
				// Adds true or false (Json booleans)
				jsonBuffer.WriteString(fmt.Sprintf(`,"ign":%v`, g.StatusIgnition))
			}
		// Ignition Alert packet ($2)
		case "$2":
			if ignition, err := strconv.ParseBool(fields[9]); err != nil {
				if DEBUGGING {
					log.Printf("Parsing error for ignition %s", fields[9])
				}
				continue
			} else {
				g.StatusIgnition = ignition
				if ignition {
					jsonBuffer.WriteString(fmt.Sprintf(`,"alert":"%s"`, "Ignition on"))
				} else {
					jsonBuffer.WriteString(fmt.Sprintf(`,"alert":"%s"`, "Ignition off"))
				}
			}
		// Main Battery Alert packet ($3)
		case "$3":
			if batConnected, err := strconv.ParseBool(fields[9]); err != nil {
				if DEBUGGING {
					log.Printf("Parsing error for batConnected %s", fields[9])
				}
				continue
			} else {
				g.StatusBattery = batConnected
				if batConnected {
					jsonBuffer.WriteString(fmt.Sprintf(`,"alert":"%s"`, "Battery connected"))
				} else {
					jsonBuffer.WriteString(fmt.Sprintf(`,"alert":"%s"`, "Battery disconnected"))
				}
			}
		// Low Battery Alert packet ($4)
		case "$4":
			jsonBuffer.WriteString(fmt.Sprintf(`,"alert":"%s"`, "Battery low"))
		// Harsh Acceleration Alert packet ($5)
		case "$5":
			jsonBuffer.WriteString(fmt.Sprintf(`,"alert":"%s"`, "Harsh Acceleration"))
		// Harsh Braking Alert packet ($6)
		case "$6":
			jsonBuffer.WriteString(fmt.Sprintf(`,"alert":"%s"`, "Harsh Braking"))
		// Overspeeding Alert packet ($7)
		case "$7":
			jsonBuffer.WriteString(fmt.Sprintf(`,"alert":"%s"`, "OverSpeeding Alert"))
			if speed, err := strconv.ParseInt(fields[9], 10, 64); err != nil {
				if DEBUGGING {
					log.Printf("Parsing error for speed %s", fields[9])
				}
				continue
			} else {
				g.Speed = int(speed)

				// Add the parsed Speed to the json
				jsonBuffer.WriteString(fmt.Sprintf(`,"speed":%d`, g.Speed))
			}
		// Box Alert packet ($8)
		case "$8":
			if boxOpen, err := strconv.ParseBool(fields[9]); err != nil {
				if DEBUGGING {
					log.Printf("Parsing error for boxOpen %s", fields[9])
				}
				continue
			} else {
				g.StatusBattery = boxOpen
				if boxOpen {
					jsonBuffer.WriteString(fmt.Sprintf(`,"alert":"%s"`, "Box Opened"))
				} else {
					jsonBuffer.WriteString(fmt.Sprintf(`,"alert":"%s"`, "Box Closed"))
				}
			}
		// SOS Alert packet ($9)
		case "$9":
			jsonBuffer.WriteString(fmt.Sprintf(`,"alert":"%s"`, "SOS"))
		}

		jsonBuffer.WriteString(`}}]}`)
		jsonString := jsonBuffer.String()
		c <- &jsonString
		// c <- g
		if DEBUGGING {
			log.Printf("Parsed dumped")
		}
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
