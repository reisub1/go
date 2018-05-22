package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"io"
	"net"
	"runtime"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func check(err error, message string) {
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s", message)
}

type Client struct {
	message string
	conn    net.Conn
}

type Packet struct {
	deviceID  string
	time      string
	latitude  string
	longitude string
	speed     string
	angle     string
	date      string
	status    string
}

func replyResponse(client chan Client) {
	for {
		ic := <-client

		for begin := time.Now(); time.Now().Sub(begin) < time.Second; {
		}

		ic.conn.Write([]byte("\n200 OK\n"))
	}
}

func parseAndInsertDB(data string, p *Packet) {

	if strings.Contains(data, "*ZJ#") {
		return
	}

	s := strings.Split(data, ",")
	k := strings.Split(s[12], "#")

	p.deviceID = s[1]
	p.time = s[3]
	p.latitude = s[5]
	p.longitude = s[7]
	p.speed = s[9]
	p.angle = s[10]
	p.date = s[11]
	p.status = k[0]

	// put user:pw@/databasename
	db, err := sql.Open("mysql", "root:gowtham@/data")

	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err.Error())
	}

	// Insert into Table ...
	result, err := db.Exec(`INSERT INTO packetTable (deviceID, time, latitude, longitude, speed, angle, date, status)
							VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, p.deviceID, p.time, p.latitude, p.longitude, p.speed, p.angle, p.date, p.status)

	if err != nil {
		println("Exec err:", err.Error())
	} else {
		id, err := result.LastInsertId()
		if err != nil {
			println("Error:", err.Error())
		} else {
			println("LastInsertId:", id)
		}
	}
}

var count = 0

func main() {

	runtime.GOMAXPROCS(4)

	clients := make(chan Client)

	go replyResponse(clients)

	// can change the port number here
	ln, err := net.Listen("tcp", ":8000")
	check(err, "Server is ready. Check if mysql.service is online.\n")

	for {
		conn, err := ln.Accept()
		count++
		check(err, "Accepted Connection")
		fmt.Printf(" %s\n", conn.RemoteAddr().String())

		go func() {
			buffer := bufio.NewReader(conn)
			for {
				message, err := buffer.ReadString('\n')
				// p := &Packet{}

				if err != nil {
					if err == io.EOF {
						continue
					}
					fmt.Printf("%s disconnected\n", conn.RemoteAddr().String())
					count--
					break
				} else {
					fmt.Printf("%s sent %s", conn.RemoteAddr().String(), message)
				}

				// tokenizes the data and inserts into the DB
				// go parseAndInsertDB(message, p)

				// send the client along the channel
				clients <- Client{message, conn}
			}
		}()
	}
}
