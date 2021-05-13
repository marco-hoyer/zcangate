package app

import (
	"github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"os"
	"time"
)

const (
	bucket = "home-metrics"
	org    = "home"
)

type Influxdb struct {
	client   influxdb2.Client
	writeAPI api.WriteAPI
}

func (i *Influxdb) Connect() {
	url := os.Getenv("INFLUXDB_URL")
	token := os.Getenv("INFLUXDB_TOKEN")
	i.client = influxdb2.NewClientWithOptions(url, token, influxdb2.DefaultOptions().SetBatchSize(20))
	i.writeAPI = i.client.WriteAPI(org, bucket)
}

func (i *Influxdb) Disconnect() {
	i.writeAPI.Flush()
	i.client.Close()
}

func (i *Influxdb) Send(measurement string, location string, unit string, identifier string, value float64) {
	p := influxdb2.NewPointWithMeasurement(measurement).
		AddTag("source", "ventilation_system").
		AddTag("location", location).
		AddTag("identifier", identifier).
		AddTag("unit", unit).
		AddField("value", value).
		SetTime(time.Now())
	i.writeAPI.WritePoint(p)
}
