package main

import (
	"github.com/nats-io/nats.go"
	"log"
	"time"
)

func main() {
	var server_ip = "0.0.0.0:4222"
	var subject = "monitor.data"
	var payload = "I am there, come to my voice!"

	opts := []nats.Option{nats.Name("Monitor")}

	// Initialise NATS-client
	nc, err := nats.Connect(server_ip, opts...)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()
	for {
		nc.Publish(subject, []byte(payload))
		nc.Flush()

		if err := nc.LastError(); err != nil {
			//TODO: catch error if NATS.server is unreachable
			log.Fatal(err)
		} else {
			log.Printf("Published [%s] : '%s'\n", subject, payload)
		}
		time.Sleep(5 * time.Second)
	}
}


