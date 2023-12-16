package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

func Ingest_MQTT(influxClient influxdb2.Client, msg mqtt.Message) {
	payload := string(msg.Payload())
	values := strings.Split(payload, ",")

	// Create a point and add to batch
	p := influxdb2.NewPointWithMeasurement("ingest")
	for i, value := range values {
		value_double, err := strconv.ParseFloat(value, 64)
		if err != nil {
			log.Println("Error converting to float ", err)
		} else {
			p.AddField(fmt.Sprintf("value%d", i), value_double)
		}

	}
	p.SetTime(time.Now())

	// Write the point to InfluxDB
	writeAPI := influxClient.WriteAPIBlocking(INFLUX_ORG, INFLUX_BUCKER)
	if err := writeAPI.WritePoint(context.Background(), p); err != nil {
		log.Println("Error writing to InfluxDB:", err)
	}
}
