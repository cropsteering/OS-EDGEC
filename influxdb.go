package main

import (
	"context"
	"fmt"
	"log"
	"slices"
	"strings"
	"time"

	"github.com/go-echarts/go-echarts/v2/opts"
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
		Influx_Disco()
		Query_Topics()
	}
}

/**
* Discovery Mode
* Search for topics to auto generate values
*
 */
func Influx_Disco() {
	var topics []string
	var found_new bool = false
	queryAPI := Influx_Client.QueryAPI(INFLUX_ORG)
	query := `from(bucket: "opensteering")
	            |> range(start: -1m)
	            |> filter(fn: (r) => r._measurement == "mqtt_consumer")
			`
	results, err := queryAPI.Query(context.Background(), query)
	if err != nil {
		log.Println(err)
	} else {
		for results.Next() {
			mqtt_topic := fmt.Sprintf("%s", results.Record().ValueByKey("topic"))
			json_array, jerr := Read_Array("topics.json")
			// This will error the very first disco
			if jerr != nil {
				log.Println("Error caching array:", jerr)
			}
			in_json := slices.Contains(json_array, mqtt_topic)
			contains := slices.Contains(topics, mqtt_topic)
			if !contains && !in_json {
				topics = append(topics, mqtt_topic)
				topic_pieces := strings.Split(mqtt_topic, "/")
				topic_user := topic_pieces[0]
				topic_zone := topic_pieces[1]
				topic_id := topic_pieces[2]
				_ = topic_user
				_ = topic_zone
				_ = topic_id
				found_new = true
			}
		}
		if found_new {
			log.Println("New topic(s) found")
			cache_err := Cache_Array(topics, "topics.json")
			if cache_err != nil {
				log.Println("Error caching array:", cache_err)
			}
		} else {
			log.Println("No new topic(s) found")
		}
		if err := results.Err(); err != nil {
			log.Println(err)
		}
	}
}

func Query_Topics() {
	cached_array, err := Read_Array("topics.json")
	if err != nil {
		log.Println("Error reading array from file:", err)
	} else {
		for _, v := range cached_array {
			queryAPI := Influx_Client.QueryAPI(INFLUX_ORG)
			query := `from(bucket: "opensteering")
					|> range(start: -1h)
					|> filter(fn: (r) => r["topic"] == "` + v + `")
					|> aggregateWindow(every: 1m, fn: mean, createEmpty: false)
					|> yield(name: "mean")`
			results, err := queryAPI.Query(context.Background(), query)
			if err != nil {
				log.Println(err)
			} else {
				var influx_data []opts.LineData
				var time_data []string
				for results.Next() {
					time_data = append(time_data, results.Record().Time().Format("2006-01-02 15:04:05"))
					influx_data = append(influx_data, opts.LineData{Value: results.Record().Value()})
				}
				MU.Lock()
				time_cache = time_data
				graph_cache[v] = influx_data
				MU.Unlock()
				if err := results.Err(); err != nil {
					log.Println(err)
				}
			}
		}
	}
}
