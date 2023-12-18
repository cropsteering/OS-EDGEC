package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

func Ingest_MQTT(influxClient influxdb2.Client, topic, payload string) {
	values := strings.Split(payload, ",")

	p := influxdb2.NewPointWithMeasurement("mqtt_data")
	p.AddTag("topic", topic)
	for i, value := range values {
		floatValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			log.Printf("Error parsing value %s as float: %v", value, err)
			continue
		}
		p.AddField(fmt.Sprintf("value%d", i), floatValue)
	}
	p.SetTime(time.Now())

	writeAPI := influxClient.WriteAPIBlocking(INFLUX_ORG, INFLUX_BUCKET)
	if err := writeAPI.WritePoint(context.Background(), p); err != nil {
		log.Println("Error writing to InfluxDB:", err)
	}
}
