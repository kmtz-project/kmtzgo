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
	var subject = "Check"
	var payload = "Hello there - reply me, please"
	opts := []nats.Option{nats.Name("command_center")}
	nc, err := nats.Connect(socket, opts...)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	msg, err := nc.Request(subject, []byte(payload), 2*time.Second)
	if err != nil {
		if nc.LastError() != nil {
			log.Fatalf("%v for request", nc.LastError())
		}
		log.Fatalf("%v for request", err)
	}
	log.Printf("Published [%s] : '%s'", subject, payload)
	log.Printf("Received  [%v] : '%s'", msg.Subject, string(msg.Data))
}
