package common

import (
	"fmt"
	"github.com/marco-hoyer/zcangate/api"
	"github.com/marco-hoyer/zcangate/can"
	"github.com/marco-hoyer/zcangate/dao"
	"github.com/tarm/serial"
	"log"
	"sync"
	"time"
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
	go func() {
		can.NewCanBusReader(s, out).Read()
	}()
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

	stateDao := dao.NewStateDao()

	log.Println("Connecting to influxdb")
	i := Influxdb{}
	i.Connect()
	defer i.Disconnect()

	w := can.CanBusWriter{Serial: s, StateDao: stateDao}

	log.Println("Starting webserver")
	runApiServer(s, &w)

	log.Println("opening CAN interface connection")
	// set CAN bus baud rate and open reading connection
	s.Write([]byte("\r\r\rC\rS2\rO\r"))
	defer s.Write([]byte("C\r"))

	//w.Write("1F015074", "84150101000000000100000003")
	//w.Write("1F035074", "84150101000000000100000003")
	//w.Write("1F055074", "84150101000000000100000003")
	//w.Write("1F075074", "84150101000000000100000003")

	log.Println("finished opening CAN interface")
	//s.Write([]byte("T1F07505180084150101000000\r"))
	//s.Write([]byte("T1F07505178100FFFFFFFF02\r"))
	//w := CanBusWriter{serial: s}
	//w.write("1F075051", "8415010100000000FFFFFFFF01")

	var wg sync.WaitGroup
	wg.Add(1)

	log.Println("reading messages")
	lines := readSerial(s)
	messages := process(lines)
	sendMeasurement(messages, i)

	wg.Wait()
}
