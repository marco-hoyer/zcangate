package app

import (
	"github.com/marco-hoyer/zcangate/api"
	"github.com/marco-hoyer/zcangate/can"
	"github.com/marco-hoyer/zcangate/dao"
	"github.com/marco-hoyer/zcangate/common"
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

func process(in <-chan can.CanBusFrame) <-chan common.Measurement {
	out := make(chan common.Measurement)
	go func() {
		for b := range in {

			//log.Println(b)
			out <- common.ToMeasurement(b)
		}
	}()
	return out
}

func sendMeasurement(in <-chan common.Measurement, i Influxdb) {
	go func() {
		for b := range in {
			if b.Name != "" {
				i.Send(b.Name, "Haus", b.Unit, "1", b.Value)
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

	var wg sync.WaitGroup
	wg.Add(1)

	log.Println("reading messages")
	lines := readSerial(s)
	messages := process(lines)
	sendMeasurement(messages, i)

	wg.Wait()
}
