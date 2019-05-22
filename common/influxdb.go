package common

import (
	"log"
	"time"
	"github.com/influxdata/influxdb/client/v2"
	"fmt"
)

const (
	url      = "http://172.17.0.1:8086"
	database = "ventilation"
	username = "root"
	password = "root"
)

type Influxdb struct {
	client client.Client
}

func (i *Influxdb) Connect() {
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     url,
		Username: username,
		Password: password,
	})
	if err != nil {
		log.Fatal(err)
	}

	i.client = c

	i.queryDB(fmt.Sprintf("CREATE DATABASE %s", database))
}

func (i *Influxdb) Disconnect() {
	i.client.Close()
}

func (i *Influxdb) queryDB(cmd string) (res []client.Result, err error) {
	q := client.Query{
		Command:  cmd,
		Database: database,
	}
	if response, err := i.client.Query(q); err == nil {
		if response.Error() != nil {
			return res, response.Error()
		}
		res = response.Results
	} else {
		return res, err
	}
	return res, nil
}

func (i *Influxdb) Send(name string, room string, unit string, identifier string, value float64) {
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  database,
		Precision: "s",
	})
	if err != nil {
		log.Fatal(err)
	}
	// Create a point and add to batch
	tags := map[string]string{
		"room":       room,
		"identifier": identifier,
		"unit":       unit,
	}

	fields := map[string]interface{}{"value": value,}

	pt, err := client.NewPoint(name, tags, fields, time.Now())
	if err != nil {
		fmt.Println(err)
	}
	bp.AddPoint(pt)

	// Write the batch
	if err := i.client.Write(bp); err != nil {
		fmt.Println(err)
	}

	// Close client resources
	if err := i.client.Close(); err != nil {
		log.Fatal(err)
	}
}
