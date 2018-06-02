// This contains a bridge which acts as a gateway(MQTT Gateway API) which interfaces with ThingsBoard's MQTT Broker
// This is to provide a compatibility layer to use legacy systems that just output telemetry data as a CSV string
// This package parses this string, converts it to json, and publishes it to the appropriate device topic on MQTT
package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"strings"
	"sync"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	logging "github.com/op/go-logging"
	gpsparser "github.com/reisub1/go/gpsAdapter/gpsparser"
	mq "github.com/reisub1/go/gpsAdapter/mq"
	evio "github.com/tidwall/evio"
)

const (
	// The IP/Port where the Legacy GPS Data is coming to
	LISTENHOST = "0.0.0.0"
	LISTENPORT = "8000"

	// The MQTT Broker IP/Port
	// MQTTHOST       = "tcp://192.168.1.20:1883"
	MQTTHOST = "tcp://127.0.0.1:1883"

	// Just a convenience variable containing Go's Example-driven method of parsing Time
	// The time and date usually comes in a concatenated format
	TIMEDATEFORMAT = "150405:020106"

	// The ThingsBoard access token for the Gateway device
	ACCESSTOKEN = "G5d65KPhrJkgL1CfflMa"
)

// This variable represents the MQTT connection that is to be persisted, and finally disconnected when the program closes
// All communication with ThingsBoard occurs through the MQTT Api
var c *mqtt.Client

// Just for convenience sake, an empty error type
var e error

// This channel contains pointers to all JSON Strings to be sent
// Another function, dispatcher, chooses one of these elements and processes it by publishing it to MQTT Broker
var jsonChan chan *string

// This is a map of string[bool] to keep track of already connected devices to prevent sending redundant connect requests
// This is synchronized with RWMutex to ensure concurrent access by different goroutines
var deviceStatus = struct {
	sync.RWMutex
	connected map[string]bool
}{connected: make(map[string]bool)}

func main() {
	f, err := os.Create("prof")
	if err != nil {
		log.Fatal(err)
	}
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()
	setUpLogging()
	signalHandler()
	log.Info("Runtime GoMAXPROCS = ", runtime.GOMAXPROCS(0))

	// Connect to the MQTT broker
	c, e = mq.Connect(MQTTHOST, ACCESSTOKEN)
	if e != nil {
		log.Critical(e)
		log.Info("Are you sure MQTT Broker is up and running?")
		os.Exit(1)
	}

	// Disconnect upon end

	jsonChan = make(chan *string, 100)
	go dispatcher(jsonChan)

	// Make an empty set of events
	var events evio.Events

	// Perform this action when the listen server on :8000 starts
	events.Serving = func(srvin evio.Server) (_ evio.Action) {
		log.Noticef("server started on port %s", LISTENPORT)
		return
	}

	// Perform this action whenever a new connection is received
	events.Opened = func(id int, info evio.Info) (_ []byte, _ evio.Options, _ evio.Action) {
		// log.Infof("Connection %d launched %s -> %s", id, info.RemoteAddr, info.LocalAddr)
		return
	}

	// Perform this action whenever new data is received
	events.Data = dataHandler

	// Perform this action whenever a connection closes
	events.Closed = func(id int, _ error) (_ evio.Action) {
		// log.Infof("Connection %d closed", id)
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

	// Assuming only messages terminated by newlines are valid
	message := strings.TrimRight(string(in), "\n")

	// Log the message for debugging
	// log.Infof("Received message of length %d: %s", len(message), message)

	// Hand off the message pointer to Parse function, it will put any parsed data it finds on the
	// parsedChannel
	go gpsparser.Parse(&message, jsonChan)

	// We are done, we can return to let the garbage collector handle this stuff
	return
}

// This is the function that dispatches goroutines to publish the newly gained data to ThingsBoard through the MQTT Gateway API
func dispatcher(workChannel chan *string) {
	for {
		select {
		case jsonString := <-workChannel:
			go func() {
				uniqid := strings.SplitN(*jsonString, "\"", 3)[1]
				deviceStatus.RLock()
				currentStatus := deviceStatus.connected[uniqid]
				deviceStatus.RUnlock()
				if !currentStatus {
					deviceJson := fmt.Sprintf(`{"device": "%s"}`, uniqid)
					mq.Publish(c, deviceJson, "v1/gateway/connect")
					deviceStatus.Lock()
					deviceStatus.connected[uniqid] = true
					deviceStatus.Unlock()
				}
				mq.Publish(c, *jsonString, "v1/gateway/telemetry")
			}()
		}
	}
}

// Global log variable to provide logging
var log = logging.MustGetLogger("server")

// In short this function just sets up colourful logging with a fixed format
func setUpLogging() {
	f, err := os.Create("/var/log/gpsAdapter.log")
	if err != nil {
		println("Unable to open log file, only logging to stderr")
	}

	logStderr := logging.NewLogBackend(os.Stderr, "", 0)
	logfile := logging.NewLogBackend(f, "", 0)
	var format = logging.MustStringFormatter(
		`%{color}%{id:0d} %{time:15:04:05.000} %{shortfunc} -> %{level:s} | %{message}%{color:reset}`,
	)
	Formatted := logging.NewBackendFormatter(logStderr, format)
	FormattedFile := logging.NewBackendFormatter(logfile, format)
	logging.SetBackend(Formatted, FormattedFile)
}

// Handle SIGINT by sending disconnect request to the MQTT Gateway Broker
func signalHandler() {
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, os.Interrupt)
	go func() {
		for sig := range sigchan {
			log.Warningf("Server closing due to Signal: %s", sig)
			// Disconnect with 1000 ms time to cleanup
			(*c).Disconnect(1000)
			// Exit cleanly
			os.Exit(0)
		}
	}()
}
