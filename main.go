package main

import (
	"github.com/tarm/serial"
	"log"
	"time"
)

func readSerial(s *serial.Port) <-chan string {
	out := make(chan string)
	newCanBusReader(s, out).Read()
	return out
}

func process(in <-chan string) <-chan Measurement {
	out := make(chan Measurement)
	go func() {
		for b := range in {

			//log.Println(b)
			out <- toMeasurement(b)
		}
	}()
	return out
}

func logLines(in <-chan Measurement) {
	go func() {
		//for b := range in {
		//	fmt.Println("Measurement: ", b)
		//}
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

	//s.Write([]byte("T1F051051485150801\r"))
	lines := readSerial(s)
	messages := process(lines)
	logLines(messages)
	time.Sleep(20 * time.Second)
}
