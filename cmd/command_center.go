package main

import (
	/*"flag"		TODO: user interaction with command center
	"os"*/
	"log"
	"time"
	"github.com/nats-io/nats.go"
)

func main() {
	var socket = "0.0.0.0:4222"
	opts := []nats.Option{nats.Name("command_center")}

	// Initialise NATS-client
	nc, err := nats.Connect(socket, opts...)
	check_err(err)
	defer nc.Close()

	var subject = "Check"
	var payload = "Hello there - reply me, please"
	request(nc, subject, payload, 2)
}

func request(nc *nats.Conn, subject string, payload string, sec_to_wait int) {
	msg, err := nc.Request(subject, []byte(payload), time.Duration(sec_to_wait)*time.Second)
	if err != nil {
		if nc.LastError() != nil {
			log.Fatalf("%v for request", nc.LastError())
		}
		log.Fatalf("%v for request", err)
	}
	log.Printf("Published [%s] : '%s'", subject, payload)
	log.Printf("Received  [%v] : '%s'", msg.Subject, string(msg.Data))
}

func check_err(e error) {
	if e != nil {
		log.Fatal(e)
	}
}