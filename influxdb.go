package main

import (
	"context"
	"log"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

var Influx_Client influxdb2.Client

func Setup_Influxdb() {
	log.Println("Starting InfluxDB")
	// Setup InfluxDB client
	Influx_Client = influxdb2.NewClient(INFLUX_URL, INFLUX_TOKEN)
	defer Influx_Client.Close()

	// Ping InfluxDB server to check status
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := Influx_Client.Ping(ctx)
	if err != nil {
		log.Println("InfluxDB not connected", err)
	} else {
		log.Println("InfluxDB connected")
	}
}
