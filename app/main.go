package app

import (
	"encoding/json"
	"fmt"
	"github.com/marco-hoyer/zcangate/api"
	"github.com/marco-hoyer/zcangate/can"
	"github.com/marco-hoyer/zcangate/common"
	"github.com/marco-hoyer/zcangate/dao"
	"github.com/streadway/amqp"
	"github.com/tarm/serial"
	"log"
	"strconv"
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

func process(in <-chan can.CanBusFrame, stateDao dao.StateDao) <-chan common.Measurement {
	out := make(chan common.Measurement)
	go func() {
		for b := range in {
			if b.PingDeviceId != 0 {
				id, _ := strconv.Atoi(fmt.Sprintf("%02x", b.PingDeviceId))
				persitedId := stateDao.GetInt(can.ComfoAirId)
				if persitedId == 0 || id < persitedId {
					stateDao.Set(can.ComfoAirId, id)
				}
			} else {
				out <- common.ToMeasurement(b)
			}
		}
	}()
	return out
}

func sendMeasurement(in <-chan common.Measurement, i Influxdb, m AmqpClient, q amqp.Queue) {
	go func() {
		for b := range in {
			if b.Name != "" {
				i.Send(b.Name, "Haus", b.Unit, "1", b.Value)
				jsonBytes, _ := json.Marshal(b)
				m.Publish(q, jsonBytes)
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

	log.Println("Connecting to RabbitMq Server")
	amqpClient := AmqpClient{}
	amqpClient.Connect()
	queue := amqpClient.QueueDeclare("ventilation/measurements")
	defer amqpClient.Disconnect()

	w := can.CanBusWriter{Serial: s, StateDao: stateDao}

	log.Println("Starting webserver")
	runApiServer(s, &w)

	log.Println("opening CAN interface connection")
	// set CAN bus baud rate and open reading connection
	s.Write([]byte("\r\r\rC\rS2\rO\r"))
	defer s.Write([]byte("C\r"))

	var wg sync.WaitGroup
	wg.Add(1)

	log.Println("reading measurements")
	canBusFrames := readSerial(s)
	measurements := process(canBusFrames, stateDao)
	sendMeasurement(measurements, i, amqpClient, queue)

	wg.Wait()
}
