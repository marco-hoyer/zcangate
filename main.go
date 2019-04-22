package main

import (
	"fmt"
	"github.com/tarm/serial"
	"log"
	"time"
)

func readSerial(s *serial.Port) <-chan CanBusFrame {
	out := make(chan CanBusFrame)
	newCanBusReader(s, out).Read()
	return out
}

func process(in <-chan CanBusFrame) <-chan Measurement {
	out := make(chan Measurement)
	go func() {
		for b := range in {

			log.Println(b)
			out <- toMeasurement(b)
		}
	}()
	return out
}

func logLines(in <-chan Measurement) {
	go func() {
		for b := range in {
			//fmt.Println("")
			if b.name != "" {
				fmt.Println("Measurement: ", b)
			}
		}
	}()
}

func main() {
	c := &serial.Config{Name: "/tmp/ttyACM0", Baud: 115200, ReadTimeout: time.Second * 5}
	s, err := serial.OpenPort(c)

	// set CAN bus baud rate and open reading connection
	s.Write([]byte("S2\r"))
	s.Write([]byte("O\r"))
	defer s.Write([]byte("C\r"))

	if err != nil {
		log.Fatal(err)
	}

	done := make(chan struct{})
	defer close(done)

	//s.Write([]byte("T1F011051485150801\r"))
	w := CanBusWriter{serial: s}
	w.write("1F07506A", "8415010100000000003C000001000000")

	//lines := readSerial(s)
	//messages := process(lines)
	//logLines(messages)
	time.Sleep(1 * time.Second)
}
