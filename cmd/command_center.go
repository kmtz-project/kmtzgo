package main

import (
	/*"flag"		TODO: user interaction with command center
	"os"*/
	"log"
	"runtime"
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

	request(nc, "custom_sensor", "publish some data for me, please", 5)

	var i = 0
	var sub_subj = "custom_sensor_data"
	nc.Subscribe(sub_subj, func(msg *nats.Msg) {
		i += 1
		printMsg(msg, i)
	})
	nc.Flush()

	err = nc.LastError()
	check_err(err)

	log.Printf("Listening on [%s]", sub_subj)
	log.SetFlags(log.LstdFlags)

	runtime.Goexit()
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

func printMsg(m *nats.Msg, i int) {
	log.Printf("[#%d] Received on [%s]: '%s'\n", i, m.Subject, string(m.Data))
}