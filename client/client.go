package main

import (
	"fmt"
	"net"
	"time"
)

var server string = "127.0.0.1:8000"

func main() {
	Conn, err := net.Dial("tcp", server)
	for err != nil {
		Conn, err = net.Dial("tcp", server)
		fmt.Println("Waiting for 5 seconds")
		time.Sleep(5 * time.Second)
	}
	s := "*ZJ,2030295119,V1,134310,A,3106.3677,N,7710.9352,E,2.38,0.00,120518,00000000#\n"

	for {
		Conn.Write([]byte(s))
		fmt.Println(s)
		fmt.Println(time.Now())
		time.Sleep(2 * time.Second)
	}
}
