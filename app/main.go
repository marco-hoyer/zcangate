package app

import (
	"github.com/marco-hoyer/zcangate/api"
	"github.com/marco-hoyer/zcangate/can"
	"github.com/marco-hoyer/zcangate/dao"
	"github.com/tarm/serial"
	"log"
	"sync"
	"time"
)

func runApiServer(serialPort *serial.Port, canBusWriter *can.BusWriter, state *dao.StateDao) {
	go func() {
		s := api.WebServer{
			SerialInterface: serialPort,
			CanBusWriter:    canBusWriter,
			State:           state,
		}
		s.Run()
	}()
}

func readSerial(s *serial.Port) <-chan can.BusFrame {
	out := make(chan can.BusFrame)
	go func() {
		can.NewCanBusReader(s, out).Read()
	}()
	return out
}

func process(in <-chan can.BusFrame) <-chan can.Measurement {
	out := make(chan can.Measurement)
	go func() {
		for b := range in {
			out <- can.ToMeasurement(b)
		}
	}()

	return out
}

func processMeasurements(in <-chan can.Measurement, influxdb Influxdb, state *dao.StateDao) {
	go func() {
		for b := range in {
			if b.Name != "" {
				influxdb.Send(b.Name, "Haus", b.Unit, "1", b.Value)
				state.Set(b.Name, b.Value)
			}
		}
	}()
}

func MainLoop() {
	portConfig := &serial.Config{Name: "/tmp/ttyACM0", Baud: 115200, ReadTimeout: time.Second * 5}
	serial, err := serial.OpenPort(portConfig)
	if err != nil {
		log.Fatal(err)
		panic(1)
	}

	log.Println("Connecting to influxdb")
	influxdb := Influxdb{}
	influxdb.Connect()
	defer influxdb.Disconnect()

	state := dao.NewStateDao()

	busWriter := can.BusWriter{Serial: serial}

	log.Println("Starting webserver")
	runApiServer(serial, &busWriter, &state)

	log.Println("opening CAN interface connection")
	// set CAN bus baud rate and open reading connection
	serial.Write([]byte("\r\r\rC\rS2\rO\r"))
	defer serial.Write([]byte("C\r"))

	var wg sync.WaitGroup
	wg.Add(1)

	log.Println("reading measurements")
	canBusFrames := readSerial(serial)
	measurements := process(canBusFrames)
	processMeasurements(measurements, influxdb, &state)

	wg.Wait()
}
