package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"

	logging "github.com/op/go-logging"
	packet "github.com/reisub1/go/newServer/packet"
	evio "github.com/tidwall/evio"
)

const (
	LISTENHOST = "0.0.0.0"
	LISTENPORT = "8000"

	SERVERHOST = "localhost"
	SERVERPORT = "80"

	MAXQUEUE = 100
)

var events evio.Events

func main() {
	setUpLogging()
	signalHandler()

	events.Serving = func(_ evio.Server) (_ evio.Action) {
		log.Noticef("server started on port %s", LISTENPORT)
		return
	}

	events.Opened = func(id int, info evio.Info) (_ []byte, _ evio.Options, _ evio.Action) {
		log.Infof("Connection %d launched %s -> %s", id, info.LocalAddr, info.RemoteAddr)
		return
	}
	events.Data = dataHandler

	events.Closed = func(id int, _ error) (_ evio.Action) {
		log.Infof("Connection %d closed", id)
		return
	}

	if err := evio.Serve(events, fmt.Sprintf("tcp://%s:%s", LISTENHOST, LISTENPORT)); err != nil {
		panic(err.Error())
	}
	defer os.Exit(0)
	log.Notice("Server closing")
}

func dataHandler(id int, in []byte) (out []byte, action evio.Action) {
	message := strings.Split(string(in), "\n")
	log.Infof("Received message of length %d: %s", len(message), message)
	return
}

var packetInterface = packet.Packet{}

var log = logging.MustGetLogger("server")

func setUpLogging() {
	logStderr := logging.NewLogBackend(os.Stderr, "", 0)
	var format = logging.MustStringFormatter(
		`%{color}%{time:15:04:05.000} %{shortfunc} -> %{level:s} %{id:0d} | %{message}%{color:reset}`,
	)
	Formatted := logging.NewBackendFormatter(logStderr, format)
	logging.SetBackend(Formatted)
}

func signalHandler() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			log.Noticef("Server closing due to Signal: %s", sig)
			os.Exit(0)
		}
	}()
}
