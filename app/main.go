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
		server := api.WebServer{
			SerialInterface: serialPort,
			CanBusWriter:    canBusWriter,
			State:           state,
		}
		server.Run()
	}()
}

func readSerial(s *serial.Port) <-chan can.BusFrame {
	out := make(chan can.BusFrame)
	go func() {
		can.NewCanBusReader(s, out).Read()
	}()
	return out
}

func process(in <-chan can.BusFrame, busWriter *can.BusWriter) <-chan can.Measurement {
	out := make(chan can.Measurement)
	go func() {
		for b := range in {
			if b.BinaryId&0xFFFFFFC0 == 0x10000000 {
				busWriter.SetDeviceID(int(b.BinaryId & 0x3f))
			} else {
				out <- can.ToMeasurement(b)
			}
		}
	}()
	return out
}

func processMeasurements(in <-chan can.Measurement, influxdb Influxdb, state *dao.StateDao) <-chan interface{} {
	heartbeatStream := make(chan interface{}, 1)

	sendHeartBeat := func() {
		select {
		case heartbeatStream <- struct{}{}:
		default:
		}
	}

	go func() {
		for b := range in {
			if b.Name != "" {
				influxdb.Send(b.Name, "Haus", b.Unit, "1", b.Value)
				state.Set(b.Name, b.Value)
				sendHeartBeat()
			}
		}
	}()

	return heartbeatStream
}

func checkHeartbeat(heartbeat <-chan interface{}) {
	go func() {
		timer := time.NewTimer(60 * time.Second)
		for {
			select {
			case <-heartbeat:
				if !timer.Stop() {
					<-timer.C
				}
				timer.Reset(60 * time.Second)
			case <-timer.C:
				log.Fatal("Process didn't receive data after 60 seconds")
			}
		}
	}()
}

func MainLoop() {
	portConfig := &serial.Config{Name: "/tmp/ttyACM0", Baud: 115200, ReadTimeout: time.Second * 5}
	serialPort, err := serial.OpenPort(portConfig)
	if err != nil {
		log.Fatal(err)
	}

	defer serialPort.Close()

	log.Println("Connecting to influxdb")
	influxdb := Influxdb{}
	influxdb.Connect()
	defer influxdb.Disconnect()

	state := dao.NewStateDao()
	busWriter := can.BusWriter{Serial: serialPort}

	log.Println("Starting webserver")
	runApiServer(serialPort, &busWriter, &state)

	log.Println("opening CAN interface connection")
	// set CAN bus baud rate and open reading connection
	serialPort.Write([]byte("\r\r\rC\rS2\rO\r"))
	defer serialPort.Write([]byte("C\r"))

	var wg sync.WaitGroup
	wg.Add(1)

	log.Println("reading measurements")
	canBusFrames := readSerial(serialPort)
	measurements := process(canBusFrames, &busWriter)
	heartbeat := processMeasurements(measurements, influxdb, &state)
	checkHeartbeat(heartbeat)

	wg.Wait()
}
