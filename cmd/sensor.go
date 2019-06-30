package main

import (
	"github.com/nats-io/nats.go"
	"log"
	"os"
	"os/signal"
	"time"
)

func main() {
	var socket = "0.0.0.0:4222"
	var queueName = "NATS-RPLY-22"
	opts := []nats.Option{nats.Name("sensor")}
	opts = setupConnOptions(opts)

	// Initialise NATS-client
	nc, err := nats.Connect(socket, opts...)
	check_err(err)

	listen_and_reply(nc, "custom_sensor", "ok! start collecting data", queueName)

	for {
		time.Sleep(2*time.Second)

		var pub_subject = "custom_sensor_data"
		var pub_payload = "custom_data inside this message"
		nc.Publish(pub_subject, []byte(pub_payload))
		nc.Flush()

		if err := nc.LastError(); err != nil {
			log.Fatal(err)
		} else {
			log.Printf("Published [%s] : '%s'\n", pub_subject, pub_payload)
		}
	}

	interrupt_handler(nc)
}

func listen_and_reply(nc *nats.Conn, subject string, payload string, queueName string) {
	var i = 0
	nc.QueueSubscribe(subject, queueName, func(msg *nats.Msg) {
		i++
		printMsg(msg, i)
		msg.Respond([]byte(payload))
	})
	nc.Flush()
	err := nc.LastError()
	check_err(err)

	log.Printf("Listening on [%s]", subject)
	log.SetFlags(log.LstdFlags)
}

func interrupt_handler(nc *nats.Conn) {
	// Setup the interrupt handler to drain so we don't miss
	// requests when scaling down.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	log.Println()
	log.Printf("Draining...")
	nc.Drain()
	log.Fatalf("Exiting")
}

func check_err(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func printMsg(m *nats.Msg, i int) {
	log.Printf("[#%d] Received on [%s]: '%s'\n", i, m.Subject, string(m.Data))
}

func setupConnOptions(opts []nats.Option) []nats.Option {
	totalWait := 10 * time.Minute
	reconnectDelay := time.Second

	opts = append(opts, nats.ReconnectWait(reconnectDelay))
	opts = append(opts, nats.MaxReconnects(int(totalWait/reconnectDelay)))
	opts = append(opts, nats.DisconnectHandler(func(nc *nats.Conn) {
		log.Printf("Disconnected: will attempt reconnects for %.0fm", totalWait.Minutes())
	}))
	opts = append(opts, nats.ReconnectHandler(func(nc *nats.Conn) {
		log.Printf("Reconnected [%s]", nc.ConnectedUrl())
	}))
	opts = append(opts, nats.ClosedHandler(func(nc *nats.Conn) {
		log.Fatalf("Exiting: %v", nc.LastError())
	}))

	return opts
}

