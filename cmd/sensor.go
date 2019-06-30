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
	var subject = "Check"
	var payload = "I am there, come to my voice!"
	var queueName = "NATS-RPLY-22"
	var i = 0

	opts := []nats.Option{nats.Name("sensor")}
	opts = setupConnOptions(opts)

	nc, err := nats.Connect(socket, opts...)
	if err != nil {
		log.Fatal(err)
	}

	nc.QueueSubscribe(subject, queueName, func(msg *nats.Msg) {
		i++
		printMsg(msg, i)
		msg.Respond([]byte(payload))
	})
	nc.Flush()
	if err := nc.LastError(); err != nil {
		log.Fatal(err)
	}

	log.Printf("Listening on [%s]", subject)
	log.SetFlags(log.LstdFlags)


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