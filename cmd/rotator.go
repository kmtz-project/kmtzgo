package main

import (
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/nats-io/nats.go"
)

func printMsg(m *nats.Msg, i int) {
	log.Printf("[#%d] Received on [%s]: '%s'\n", i, m.Subject, string(m.Data))
}

func main() {
	var url = "0.0.0.0:4222"
	var queueName = "ROTATOR-RPLY"
	var subj = "rotate"
	var msgCnt = 0

	// Connect Options.
	opts := []nats.Option{nats.Name("Rotator Responder")}
	opts = setupConnOptions(opts)

	// Connect to NATS
	nc, err := nats.Connect(url, opts...)
	if err != nil {
		log.Fatal(err)
	}

	nc.QueueSubscribe(subj, queueName, func(msg *nats.Msg) {
		msgCnt++
		printMsg(msg, msgCnt)
		msg.Respond([]byte(generateRespond()))
	})
	nc.Flush()

	if err := nc.LastError(); err != nil {
		log.Fatal(err)
	}

	// Enable log showTimestamps
	log.SetFlags(log.LstdFlags)

	log.Printf("Listening on [%s]", subj)


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

func generateRespond() string {
	return "myreply"
}

func setupConnOptions(opts []nats.Option) []nats.Option {
	totalWait := 10 * time.Minute
	reconnectDelay := time.Second

	opts = append(opts, nats.ReconnectWait(reconnectDelay))
	opts = append(opts, nats.MaxReconnects(int(totalWait/reconnectDelay)))
	opts = append(opts, nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
		log.Printf("Disconnected due to: %s, will attempt reconnects for %.0fm", err, totalWait.Minutes())
	}))
	opts = append(opts, nats.ReconnectHandler(func(nc *nats.Conn) {
		log.Printf("Reconnected [%s]", nc.ConnectedUrl())
	}))
	opts = append(opts, nats.ClosedHandler(func(nc *nats.Conn) {
		log.Fatalf("Exiting: %v", nc.LastError())
	}))
	return opts
}
