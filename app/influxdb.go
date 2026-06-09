package app

import (
	"github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"time"
)

type Influxdb struct {
	client   influxdb2.Client
	writeAPI api.WriteAPI
	location string
}

func (i *Influxdb) Connect(cfg Config) {
	i.client = influxdb2.NewClientWithOptions(cfg.InfluxDBURL, cfg.InfluxDBToken,
		influxdb2.DefaultOptions().SetBatchSize(20))
	i.writeAPI = i.client.WriteAPI(cfg.InfluxDBOrg, cfg.InfluxDBBucket)
	i.location = cfg.InfluxDBLocation
}

func (i *Influxdb) Disconnect() {
	i.writeAPI.Flush()
	i.client.Close()
}

func (i *Influxdb) Send(measurement string, unit string, identifier string, value float64) {
	p := influxdb2.NewPointWithMeasurement(measurement).
		AddTag("source", "ventilation_system").
		AddTag("location", i.location).
		AddTag("identifier", identifier).
		AddTag("unit", unit).
		AddField("value", value).
		SetTime(time.Now())
	i.writeAPI.WritePoint(p)
}
