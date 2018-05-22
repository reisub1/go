package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"time"

	logging "github.com/op/go-logging"
)

const (
	HOST = "localhost"
	PORT = "8000"
)

func main() {
	setUpLogging()
	listener, err := net.Listen("tcp", ":8000")
	defer listener.Close()
	if err != nil {
		log.Critical("Failed to listen on port 8000")
		panic(err)
	}
	log.Info("Sucessfully set up socket")

	for {
		client, err := listener.Accept()
		if err != nil {
			log.Critical("Could not accept connection (Memory overflow?)")
		}
		go getMessage(client)
	}
}

func clientHandler(client) error {

}

func getMessage(conn) {
	buffer := bufio.NewReader(conn)

	select {
	case msg, err := <-buffer.ReadString('\n'):
		if err != nil {
		}
	case <-time.After(30 * time.Second):
		fmt.Println("timeout 30")

	}
}
func jsonifyAndSend(msg string) {

}

var log = logging.MustGetLogger("server")

func setUpLogging() {
	logStderr := logging.NewLogBackend(os.Stderr, "", 0)
	var format = logging.MustStringFormatter(
		`%{color}%{time:15:04:05.000} %{shortfunc} -> %{level:s} %{id:0d} |%{color:reset} %{message}`,
	)
	Formatted := logging.NewBackendFormatter(logStderr, format)
	logging.SetBackend(Formatted)
}

type packet struct {
	deviceID  string
	time      string
	latitude  string
	longitude string
	speed     string
	angle     string
	date      string
	status    string
}

var packetInterface = packet{}
