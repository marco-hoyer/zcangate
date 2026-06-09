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

func runApiServer(canBusWriter *can.BusWriter, state *dao.StateDao, cfg Config) {
	go func() {
		server := api.WebServer{
			CanBusWriter: canBusWriter,
			State:        state,
			Addr:         cfg.HTTPAddr,
			AuthToken:    cfg.CommandAuthToken,
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
				influxdb.Send(b.Name, b.Unit, "1", b.Value)
				state.Set(b.Name, b.Value)
				sendHeartBeat()
			}
		}
	}()

	return heartbeatStream
}

func checkHeartbeat(heartbeat <-chan interface{}, timeout time.Duration) {
	go func() {
		timer := time.NewTimer(timeout)
		for {
			select {
			case <-heartbeat:
				if !timer.Stop() {
					<-timer.C
				}
				timer.Reset(timeout)
			case <-timer.C:
				log.Fatalf("no measurement received within %s", timeout)
			}
		}
	}()
}

func MainLoop() {
	cfg := LoadConfig()

	portConfig := &serial.Config{
		Name:        cfg.SerialPort,
		Baud:        cfg.SerialBaud,
		ReadTimeout: time.Second * 5,
	}
	serialPort, err := serial.OpenPort(portConfig)
	if err != nil {
		log.Fatal(err)
	}
	defer serialPort.Close()

	log.Println("connecting to InfluxDB")
	influxdb := Influxdb{}
	influxdb.Connect(cfg)
	defer influxdb.Disconnect()

	state := dao.NewStateDao()
	busWriter := can.BusWriter{Serial: serialPort}

	log.Println("starting web server on", cfg.HTTPAddr)
	runApiServer(&busWriter, &state, cfg)

	log.Println("opening CAN interface connection")
	serialPort.Write([]byte("\r\r\rC\rS2\rO\r"))
	defer serialPort.Write([]byte("C\r"))

	var wg sync.WaitGroup
	wg.Add(1)

	log.Println("reading measurements")
	canBusFrames := readSerial(serialPort)
	measurements := process(canBusFrames, &busWriter)
	heartbeat := processMeasurements(measurements, influxdb, &state)
	checkHeartbeat(heartbeat, cfg.HeartbeatTimeout)

	wg.Wait()
}
