package main

import (
	"fmt"
	"net"
	"time"
)

var server string = "127.0.0.1:8000"

var TIMEFORMAT = "150405"
var DATEFORMAT = "020106"

func main() {
	Conn, err := net.Dial("tcp", server)
	for err != nil {
		Conn, err = net.Dial("tcp", server)
		fmt.Println("Waiting for 5 seconds")
		time.Sleep(5 * time.Second)
	}

	for {
		s := fmt.Sprintf("*ZJ,2030295119,V1,%s,A,3106.3677,N,7710.9352,E,2.38,0.00,%s,00000000#\n", time.Now().Format(TIMEFORMAT), time.Now().Format(DATEFORMAT))
		Conn.Write([]byte(s))
		fmt.Println(s)
		// fmt.Println(time.Now())
		time.Sleep(2 * time.Second)
	}
}
