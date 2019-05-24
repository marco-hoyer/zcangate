package common

import (
	"github.com/tarm/serial"
	"log"
	"time"
	"sync"
	"fmt"
	"github.com/marco-hoyer/zcangate/api"
	"github.com/marco-hoyer/zcangate/can"
)

func runApiServer(s *serial.Port, w *can.CanBusWriter) {
	go func() {
		s := api.WebServer{
			s,
			w,
		}
		s.Run()
	}()
}

func readSerial(s *serial.Port) <-chan can.CanBusFrame {
	out := make(chan can.CanBusFrame)
	can.NewCanBusReader(s, out).Read()
	return out
}

func process(in <-chan can.CanBusFrame) <-chan Measurement {
	out := make(chan Measurement)
	go func() {
		for b := range in {

			//log.Println(b)
			out <- ToMeasurement(b)
		}
	}()
	return out
}

func sendMeasurement(in <-chan Measurement, i Influxdb) {
	go func() {
		for b := range in {
			//fmt.Println("")
			if b.name != "" {
				i.Send(b.name, "Haus", b.unit, "1", b.value)
				fmt.Println("Measurement: ", b)
			}
		}
	}()
}



func MainLoop() {
	c := &serial.Config{Name: "/tmp/ttyACM0", Baud: 115200, ReadTimeout: time.Second * 5}
	s, err := serial.OpenPort(c)
	if err != nil {
		log.Fatal(err)
		panic(1)
	}

	log.Println("Connecting to influxdb")
	i := Influxdb{}
	i.Connect()
	defer i.Disconnect()

	w := can.CanBusWriter{Serial: s}

	log.Println("Starting webserver")
	runApiServer(s, &w)

	log.Println("opening CAN interface connection")
	// set CAN bus baud rate and open reading connection
	s.Write([]byte("S2\r"))
	s.Write([]byte("O\r"))
	defer s.Write([]byte("C\r"))

	log.Println("finished opening CAN interface")
	//s.Write([]byte("T1F07505180084150101000000\r"))
	//s.Write([]byte("T1F07505178100FFFFFFFF02\r"))
	//w := CanBusWriter{serial: s}
	//w.write("1F075051", "8415010100000000FFFFFFFF01")
	//w.write("1F035057", "8415010100000000FFFFFFFF03")


	var wg sync.WaitGroup
	wg.Add(1)

	log.Println("reading messages")
	lines := readSerial(s)
	messages := process(lines)
	sendMeasurement(messages, i)

	wg.Wait()
}
