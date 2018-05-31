// This contains a bridge which acts as a gateway(MQTT Gateway API) which interfaces with ThingsBoard's MQTT Broker
// This is to provide a compatibility layer to use legacy systems that just output telemetry data as a CSV string
// This package parses this string, converts it to json, and publishes it to the appropriate device topic on MQTT
package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	logging "github.com/op/go-logging"
	gpsparser "github.com/reisub1/go/gpsAdapter/gpsparser"
	mq "github.com/reisub1/go/gpsAdapter/mq"
	evio "github.com/tidwall/evio"
)

const (
	LISTENHOST = "0.0.0.0"
	LISTENPORT = "8000"

	// MQTTHOST       = "tcp://192.168.1.20:1883"
	MQTTHOST       = "tcp://127.0.0.1:1883"
	TIMEDATEFORMAT = "150405:020106"
	ACCESSTOKEN    = "G5d65KPhrJkgL1CfflMa"
)

var events evio.Events
var c *mqtt.Client
var e error
var parsedChan chan *gpsparser.GPSParsed

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
	c, e = mq.Connect(MQTTHOST, ACCESSTOKEN)
	if e != nil {
		log.Critical(e)
		log.Info("Are you sure MQTT Broker is up and running?")
		os.Exit(1)
	}
	// Disconnect upon end

	parsedChan = make(chan *gpsparser.GPSParsed, 100)
	go dispatcher(parsedChan)

	// Perform this action when the listen server on :8000 starts
	events.Serving = func(srvin evio.Server) (_ evio.Action) {
		log.Noticef("server started on port %s", LISTENPORT)
		return
	}

	// Perform this action whenever a new connection is received
	events.Opened = func(id int, info evio.Info) (_ []byte, _ evio.Options, _ evio.Action) {
		log.Infof("Connection %d launched %s -> %s", id, info.RemoteAddr, info.LocalAddr)
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
// This parses the message in the WTD Protocol format and then Publishes it as JSON for the thingsboard MQTT Gateway API
func dataHandler(id int, in []byte) (out []byte, action evio.Action) {
	message := strings.Split(string(in), "\n")[0]
	log.Infof("Received message of length %d: %s", len(message), message)

	go gpsparser.Parse(&message, parsedChan)

	return
}

func dispatcher(workChannel chan *gpsparser.GPSParsed) {
	for {
		select {
		case work := <-workChannel:
			go func() {
				deviceStatus.RLock()
				currentStatus := deviceStatus.connected[work.Uniqid]
				deviceStatus.RUnlock()
				if !currentStatus {
					deviceJson := fmt.Sprintf(`{"device": "%s"}`, work.Uniqid)
					mq.Publish(c, deviceJson, "v1/gateway/connect")
					deviceStatus.Lock()
					deviceStatus.connected[work.Uniqid] = true
					deviceStatus.Unlock()
				}
				telemetryJson := fmt.Sprintf(`{"%s":[{"ts":%d,"values":{"lat":%f,"lon":%f}}]}`,
					work.Uniqid, work.TS_Millis, work.ActualLat, work.ActualLng)
				mq.Publish(c, telemetryJson, "v1/gateway/telemetry")
			}()
		}
	}
}

// Global log variable to provide logging
var log = logging.MustGetLogger("server")

func setUpLogging() {
	verbosePtr := flag.Bool("v", false, "Verbose output")
	flag.Parse()
	f, err := os.Create("/var/log/tbBridge.log")
	if err != nil {
		println("Unable to open log file, only logging to stderr")
	}

	if !*verbosePtr {
		logStderr := logging.NewLogBackend(os.Stderr, "", 0)
		logfile := logging.NewLogBackend(f, "", 0)
		var format = logging.MustStringFormatter(
			`%{color}%{id:0d} %{time:15:04:05.000} %{shortfunc} -> %{level:s} | %{message}%{color:reset}`,
		)
		Formatted := logging.NewBackendFormatter(logStderr, format)
		FormattedFile := logging.NewBackendFormatter(logfile, format)
		logging.SetBackend(Formatted, FormattedFile)
	}
}

// Handle SIGINT by sending disconnect request to the API
func signalHandler() {
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, os.Interrupt)
	go func() {
		for sig := range sigchan {
			log.Warningf("Server closing due to Signal: %s", sig)
			(*c).Disconnect(1000)
			os.Exit(0)
		}
	}()
}
