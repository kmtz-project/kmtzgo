package main

import (
	"github.com/nats-io/nats.go"
	"log"
	"strings"
	"time"
	"os/exec"
)

func main() {
	var subject = "monitor.measure_temp"
	var server_ip = "0.0.0.0:4222"

	opts := []nats.Option{nats.Name("monitor")}

	// Initialise NATS-client
	nc, err := nats.Connect(server_ip, opts...)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()
	for {
		monitor_app := "/opt/vc/bin/vcgencmd"
		monitor_cmds := [2]string{"measure_volts", "measure_temp"}
		for _, m_cmd := range monitor_cmds {
			sh_cmd := exec.Command(monitor_app, m_cmd)
			stdout, _ := sh_cmd.Output()
			stdout = []byte(strings.TrimSuffix(string(stdout), "\n"))

			nc.Publish(subject, []byte(stdout))
			nc.Flush()

			if err := nc.LastError(); err != nil {
				log.Fatal(err)
			} else {
				log.Printf("Published [%s] : '%s'\n", subject, stdout)
			}
		}


		time.Sleep(1 * time.Second)
	}
}


