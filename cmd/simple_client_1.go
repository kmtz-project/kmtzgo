// main_client_1
package main

import (
	"fmt"
	//"log"
	//"strconv"
	//"strings"
	//"sync"
	"os"
	"time"

	"github.com/nats-io/nats.go"
)

func main() {
	fmt.Println("Hello World! I am First Client!")

	// Connect to a server
	nc, _ := nats.Connect(nats.DefaultURL)
	argsWithoutProg := os.Args[1:]

	for i := 0; i < 2; i++ {
		// Simple Async Subscriber
		//nc.Subscribe("client2", func(m *nats.Msg) {
		//	fmt.Printf("Received a message: %s\n", string(m.Data))
		//})

		// Simple Publisher
		nc.Publish("Check", []byte(argsWithoutProg[0]))
		fmt.Printf("i send " + argsWithoutProg[0] + "\n")
		i++

		timer := time.NewTimer(time.Millisecond * 100)
		<-timer.C

	}
}
