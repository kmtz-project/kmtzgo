package main

import (
	//"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"flag"
	"log"
	"os"
	"os/signal"

	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/goburrow/serial"
	"github.com/nats-io/nats.go"
)

type commandFromJSON struct {
	CommandForStm32 string
	Arguments       string
}

func main() {
	var socket = "0.0.0.0:4222"
	var subject = "Check"
	var payload = "I am there, come to my voice!"
	//var queueName = "NATS-RPLY-22"
	opts := []nats.Option{nats.Name("sensor")}
	opts = setupConnOptions(opts)

	// Initialise NATS-client
	nc, err := nats.Connect(socket, opts...)
	check_err(err)

	port := connect_port()

	listen_and_reply(nc, subject, payload /*, queueName*/, port)

	interrupt_handler(nc)
}

var (
	address  string
	baudrate int
	databits int
	stopbits int
	parity   string

	message string
)

var encoder uint16
var receiveEncoder bool

type bitString string

func search_in_data(port serial.Port, data []byte, nc *nats.Conn) {
	str := string(data)

	log.Println("i receive:" + str)
	if strings.HasPrefix(str, "special:") {
		str_split := strings.Split(str, "special:")
		data_to_uart, err := hex.DecodeString(str_split[1])
		if err != nil {
			panic(err)
		}
		port_write(port, data_to_uart)
	} else {
		var strCom string
		var strComArg string
		strComArg = ""
		if strings.HasPrefix(str, "enturn") {
			strCom = "enturn"
			fmt.Printf("t and en\n")
		} else if strings.HasPrefix(str, "turn") {
			strCom = "turn"
			fmt.Printf("t\n")
		} else if strings.HasPrefix(str, "encoder") {
			strCom = "encoder"
			fmt.Printf("r\n")
		} else if strings.HasPrefix(str, "led") {
			strCom = "led"
			fmt.Printf("l\n")
		}

		if strCom != "encoder" {
			str_split := strings.Split(str, strCom+":")
			strComArg = str_split[1]
			fmt.Printf("strComArg %s\n", strComArg)
		}

		jsonstr := map[string]string{"CommandForStm32": strCom, "Arguments": strComArg}
		comJson, _ := json.Marshal(jsonstr)

		var myCommand commandFromJSON

		json.Unmarshal([]byte(comJson), &myCommand)

		var lengthOfMessage uint8
		var numberOfMessage uint8
		var commandOfMessage uint8
		var crcOfMessage uint8
		var turnIterator int
		var turnAndEncoder bool
		turnAndEncoder = false
		turnIterator = 0
		numberOfMessage = 0
		lengthOfMessage = 2
		commandOfMessage = 0

		if strings.Compare(myCommand.CommandForStm32, "enturn") == 0 {
			commandOfMessage = 0x40
			turnAndEncoder = true
			fmt.Printf("and encoder\n")
			myCommand.CommandForStm32 = "turn"
		}

		if strings.Compare(myCommand.CommandForStm32, "turn") == 0 {
			commandOfMessage |= 0x02
			commandOfMessage |= 0x04 //2 MS3
			commandOfMessage |= 0x08 //3 MS2
			commandOfMessage |= 0x10 //4 MS1

			turnIterator, _ = strconv.Atoi(strComArg)
			fmt.Printf("turnIterator %d\n", turnIterator)
			if turnIterator < 0 {
				turnIterator = -turnIterator
				commandOfMessage |= 0x01
			}
			fmt.Printf("only turn\n")

			buf := []byte{lengthOfMessage, numberOfMessage, commandOfMessage}
			crcOfMessage = crc8(buf, 3)
			data_to_uart := []byte{lengthOfMessage, numberOfMessage, commandOfMessage, crcOfMessage}

			if turnIterator > 0 {
				for i := 0; i < turnIterator; i++ {
					fmt.Printf("turnIterator: %d %d\n", turnIterator, i)
					port_write(port, data_to_uart)

					if turnAndEncoder == true {
						receiveEncoder = true
						//port_read(port)
						data2 := make([]byte, 5)
						n, err := port.Read(data2)

						if err != nil {
							log.Fatal(err)
							fmt.Printf("error port read")
						}
						if n > 0 {
							fmt.Printf("n: %d\n", n)
							encoder = uint16(data2[2]) << 8
							encoder = encoder + uint16(data2[3])
							fmt.Println("encoder: ", encoder)

							dataEncoder := make([]byte, 2)
							dataEncoder[0] = data2[2]
							dataEncoder[1] = data2[3]
							nc.Publish("Encoder", []byte(dataEncoder))
							//nc.Flush()
						}
						data2 = nil
						timer := time.NewTimer(time.Millisecond * 10)
						<-timer.C
						//port.Read(data2)

						//data = data[0:n]
						//port.Read(readPortData)
					} else {
						timer := time.NewTimer(time.Millisecond * 200)
						<-timer.C
					}

				}
			}

		} else if strings.Compare(myCommand.CommandForStm32, "encoder") == 0 {
			fmt.Printf("only encoder\n")
			commandOfMessage = 0x40
			buf := []byte{lengthOfMessage, numberOfMessage, commandOfMessage}
			crcOfMessage = crc8(buf, 3)
			data_to_uart := []byte{lengthOfMessage, numberOfMessage, commandOfMessage, crcOfMessage}
			port_write(port, data_to_uart)
			receiveEncoder = true
			port_read(port)

		} else if strings.Compare(myCommand.CommandForStm32, "led") == 0 {
			fmt.Printf("led\n")
			commandOfMessage = 0
			commandOfMessage |= 0x80

			if strings.HasPrefix(strComArg, "on") {
				commandOfMessage |= 0x08 //if on
			}

			//commandOfMessage |= 0x03
			var ledChange uint8
			if strings.Contains(strComArg, "1") {
				ledChange = 1
			}
			if strings.Contains(strComArg, "2") {
				ledChange = 2
			}
			if strings.Contains(strComArg, "3") {
				ledChange = 3
			}

			switch ledChange {
			case 1:
				commandOfMessage |= 0x01
			case 2:
				commandOfMessage |= 0x02
			case 3:
				commandOfMessage |= 0x03
			}

			buf := []byte{lengthOfMessage, numberOfMessage, commandOfMessage}
			crcOfMessage = crc8(buf, 3)
			data_to_uart := []byte{lengthOfMessage, numberOfMessage, commandOfMessage, crcOfMessage}
			fmt.Print("send leds command\n")
			port_write(port, data_to_uart)
		}

		//mapD := map[string]int{"apple": 5, "lettuce": 7}
		//mapB, _ := json.Marshal(mapD)

		//mapD := map[string]string{"CommandForStm32": "led", "Arguments": "back 360"}
		//comJson, _ := json.Marshal(mapD)
		//log.Println("try again")
	}
}

func crc8(mas []byte, len int) byte {
	var crc byte = 0xFF
	var iter byte = 0
	for len > 0 {
		len--

		crc ^= mas[iter]
		iter++

		for i := 0; i < 8; i++ {
			if crc&0x80 > 0 {
				crc = (crc << 1) ^ 0x31
			} else {
				crc = crc << 1
			}
		}
	}
	return crc
}

func port_read(port serial.Port) ([]byte, int) {
	data := make([]byte, 5)
	n, err := port.Read(data)
	if err != nil {
		log.Fatal(err)
	}
	if n > 0 {
		//fmt.Println("receive bytes from port\n", n)
		//fmt.Println(data[0], data[1], data[2], data[3], data[4])
		data[0] = data[2]
		data[1] = data[3]
		encoder = uint16(data[2]) << 8
		encoder = encoder + uint16(data[3])
		fmt.Println("encoder: ", encoder)
	} else {
		fmt.Print("zero n")
	}

	//fmt.Println(string(data[:n]))
	return data, n
}

func port_write(port serial.Port, data []byte) {
	_, err := port.Write(data)
	if err != nil {
		log.Fatal(err)
		fmt.Printf("error port write")
	}
}

func connect_port() serial.Port {
	flag.StringVar(&address, "a", "/dev/serial0", "address")
	flag.IntVar(&baudrate, "b", 115200, "baud rate")
	flag.IntVar(&databits, "d", 8, "data bits")
	flag.IntVar(&stopbits, "s", 1, "stop bits")
	flag.StringVar(&parity, "p", "N", "parity (N/E/O)")
	flag.StringVar(&message, "m", "serial", "message")
	flag.Parse()

	config := serial.Config{
		Address:  address,
		BaudRate: baudrate,
		DataBits: databits,
		StopBits: stopbits,
		Parity:   parity,
		//Timeout:  30 * time.Second,
	}

	log.Println("connecting ", config)

	port, err := serial.Open(&config)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("connected")

	return port
}

func listen_and_reply(nc *nats.Conn, subject string, payload string /*, queueName string*/, port serial.Port) {
	log.Printf("lis\n")
	var i = 0
	//nc.QueueSubscribe(subject, queueName, func(msg *nats.Msg) {
	nc.Subscribe(subject /*, queueName*/, func(msg *nats.Msg) {
		i++
		printMsg(msg, i)
		data_to_uart := msg.Data
		receiveEncoder = false
		search_in_data(port, data_to_uart, nc)
		fmt.Printf("after search\n")
		///msg.Respond([]byte(payload))
	})
	//nc.Flush()
	if err := nc.LastError(); err != nil {
		log.Fatal(err)
	}

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

