package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

func Ingest_MQTT(influxClient influxdb2.Client, topic, payload string) {
	topic_type := strings.Split(topic, "/")
	topic_len := len(topic_type)
	if topic_len >= 4 {
		switch topic_type[3] {
		case "control":
			// Ignore
		case "config":
			// Ignore
		default:
			ingest_values(influxClient, topic, payload)
		}
	} else {
		ingest_values(influxClient, topic, payload)
	}
}

func ingest_values(influxClient influxdb2.Client, topic, payload string) {
	values := strings.Split(payload, ",")

	p := influxdb2.NewPointWithMeasurement(INFLUX_MEASUREMENT)
	p.AddTag("topic", topic)
	for i, value := range values {
		if Is_Float(value) {
			floatValue, err := strconv.ParseFloat(value, 64)
			if err != nil {
				R_LOG("Value " + value + "/" + err.Error())
			} else {
				p.AddField(fmt.Sprintf("value%d", i), floatValue)
			}
		} else {
			p.AddField("state", value)
		}
	}
	p.SetTime(time.Now())

	writeAPI := influxClient.WriteAPIBlocking(INFLUX_ORG, INFLUX_BUCKET)
	if err := writeAPI.WritePoint(context.Background(), p); err != nil {
		R_LOG("Error writing to InfluxDB: " + err.Error())
	}
}
