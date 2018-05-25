// This contains a bridge which acts as a gateway(MQTT Gateway API) which interfaces with ThingsBoard's MQTT Broker
// This is to provide a compatibility layer to use legacy systems that just output telemetry data as a CSV string
// This package parses this string, converts it to json, and publishes it to the appropriate device topic on MQTT
package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	logging "github.com/op/go-logging"
	mqtt "github.com/reisub1/go/newServer/mqtt"
	mqService "github.com/surgemq/surgemq/service"
	evio "github.com/tidwall/evio"
	"sync"
)

const (
	LISTENHOST = "0.0.0.0"
	LISTENPORT = "8000"

	// MQTTHOST       = "tcp://192.168.1.20:1883"
	MQTTHOST       = "tcp://192.168.1.20:1883"
	TIMEDATEFORMAT = "150405:020106"
	ACCESSTOKEN    = "x8AeRVAGkDh1s8iK0o97"
)

var events evio.Events
var c *mqService.Client
var e error

// This is a map of string[bool] to keep track of already connected devices to prevent sending redundant connect requests
// This is synchronized with RWMutex to ensure concurrent access by different goroutines
var deviceStatus = struct {
	sync.RWMutex
	connected map[string]bool
}{connected: make(map[string]bool)}

func main() {
	setUpLogging()
	signalHandler()

	// Connect to the MQTT broker
	c, e = mqtt.Connect(MQTTHOST, ACCESSTOKEN)
	if e != nil {
		log.Critical(e)
		log.Notice("Are you sure MQTT Broker is up and running?")
		os.Exit(1)
	}

	// Perform this action when the listen server on :8000 starts
	events.Serving = func(srvin evio.Server) (_ evio.Action) {
		log.Noticef("server started on port %s", LISTENPORT)
		return
	}

	// Perform this action whenever a new connection is received
	events.Opened = func(id int, info evio.Info) (_ []byte, _ evio.Options, _ evio.Action) {
		log.Infof("Connection %d launched %s -> %s", id, info.LocalAddr, info.RemoteAddr)
		return
	}

	// Perform this action whenever new data is received
	events.Data = dataHandler

	// Perform this action whenever a connection closes
	events.Closed = func(id int, _ error) (_ evio.Action) {
		log.Infof("Connection %d closed", id)
		return
	}

	// Start the server on LISTENHOST:LISTENPORT
	if err := evio.Serve(events, fmt.Sprintf("tcp://%s:%s", LISTENHOST, LISTENPORT)); err != nil {
		panic(err.Error())
	}
}

// dataHandler is the function called asynchronously upon a new Data connection from a client
// This parses the message in the WTD Protocol format and then outputs it as JSON for the thingsboard MQTT Gateway API
func dataHandler(id int, in []byte) (out []byte, action evio.Action) {
	message := strings.Split(string(in), "\n")[0]
	log.Infof("Received message of length %d: %s", len(message), message)
	if strings.Contains(message, "*ZJ#") {
		return
	}

	tokens := strings.Split(message, ",")
	if len(tokens) == 1 {
		log.Debugf("Invalid message(%d) received: %s", len(message), message)
		return
	}

	uniqid := tokens[1]
	deviceStatus.RLock()
	currentStatus := deviceStatus.connected[uniqid]
	deviceStatus.RUnlock()
	if !currentStatus {
		deviceJson := fmt.Sprintf(`{"device": "%s"}`, uniqid)
		mqtt.Publish(c, deviceJson, "v1/gateway/connect")
		deviceStatus.Lock()
		deviceStatus.connected[uniqid] = true
		deviceStatus.Unlock()
	}
	lat := 0.0
	if lat, e = strconv.ParseFloat(string(tokens[5]), 32); e != nil {
		lat = 0.0
		log.Errorf("Parsing error for latitude %f", lat)
	}
	// 7th field contains latitude direction information
	if strings.Compare(tokens[6], "S") == 0 {
		lat = -1 * lat
	}
	lng := 0.0
	if lng, e = strconv.ParseFloat(string(tokens[7]), 32); e != nil {
		lng = 0.0
		log.Errorf("Parsing error for longitude %f", lng)
	}
	// 9th field contains longitude direction information
	if strings.Compare(tokens[9], "W") == 0 {
		lng = -1 * lng
	}

	timeDateString := strings.Join([]string{tokens[3], tokens[11]}, ":")
	timestamp, e := time.Parse(TIMEDATEFORMAT, timeDateString)
	if e != nil {
		log.Error("Parsing error for ", timeDateString)
	}
	// Thingsboard expects time in miilis since epoch
	_ = timestamp
	// ts := timestamp.Unix() * 1000
	ts := time.Now().Unix() * 1000

	telemetryJson := fmt.Sprintf(`{"%s":[{"ts":%d,"values":{"lat":%f,"lng":%f}}]}`,
		uniqid, ts, lat/100, lng/100)
	mqtt.Publish(c, telemetryJson, "v1/gateway/telemetry")

	return
}

// Global log variable to provide logging
var log = logging.MustGetLogger("server")

func setUpLogging() {
	verbosePtr := flag.Bool("v", false, "Verbose output")
	flag.Parse()

	if !*verbosePtr {
		logStderr := logging.NewLogBackend(os.Stderr, "", 0)
		var format = logging.MustStringFormatter(
			`%{color}%{id:0d} %{time:15:04:05.000} %{shortfunc} -> %{level:s} | %{message}%{color:reset}`,
		)
		Formatted := logging.NewBackendFormatter(logStderr, format)
		logging.SetBackend(Formatted)
	}
}

// Handle SIGINT by sending disconnect requests to the API
func signalHandler() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			log.Warningf("Server closing due to Signal: %s", sig)
			os.Exit(0)
		}
	}()
}

func StupidFunction(a string) {
	fmt.Println(a)
}
