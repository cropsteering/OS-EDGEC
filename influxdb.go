package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"slices"
	"time"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

var Influx_Client influxdb2.Client
var graph_cache = make(map[string]interface{})

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
	query := `from(bucket: "` + INFLUX_BUCKET + `")
	            |> range(start: -1h)
	            |> filter(fn: (r) => r._measurement == "` + INFLUX_MEASUREMENT + `")
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

/**
* Query MQTT topics in Influx
* Then generate chart data, save to html
*
 */
func Query_Topics() {
	cached_array, err := Read_Array("topics.json")
	if err != nil {
		log.Println("Error reading array from file:", err)
	} else {
		f, err := os.Create("lineChart.html")
		if err != nil {
			log.Println(err)
		}
		for _, v := range cached_array {
			queryAPI := Influx_Client.QueryAPI(INFLUX_ORG)
			fluxQuery := `
				from(bucket: "` + INFLUX_BUCKET + `")
				|> range(start: -1h)
				|> filter(fn: (r) => r["topic"] == "` + v + `")
				|> aggregateWindow(every: 1m, fn: mean, createEmpty: false)
				|> yield(name: "mean")
			`
			results, err := queryAPI.Query(context.Background(), fluxQuery)
			if err != nil {
				log.Println(err)
			} else {
				var influx_data []opts.LineData
				var time_data []string
				var value_names []string
				var last_value string
				var v_name string

				for results.Next() {
					v_name = results.Record().Field()
					if err != nil {
						log.Println("Value name error ", err)
					} else {
						contains := slices.Contains(value_names, v_name)
						if !contains {
							log.Println("Adding value name:", v_name)
							value_names = append(value_names, v_name)
						}
					}
					if last_value == v_name {
						time_data = append(time_data, results.Record().Time().Format("2006-01-02 15:04:05"))
						influx_data = append(influx_data, opts.LineData{Value: results.Record().Value()})
					} else {
						influx_data = nil
						time_data = nil
					}
					graph_cache["name"] = fmt.Sprintf("%s", results.Record().ValueByKey("topic"))
					graph_cache[v_name] = influx_data
					last_value = v_name
				}

				graph_cache["times"] = time_data

				if results.Err() != nil {
					log.Printf("Query processing error: %v\n", results.Err().Error())
				}

				json_obj, _ := json.MarshalIndent(graph_cache, "", "  ")
				value_count, _ := array_count(json_obj)
				value_count = value_count - 1

				line := charts.NewLine()
				line.SetGlobalOptions(
					charts.WithTooltipOpts(opts.Tooltip{
						Show:      true,
						Trigger:   "axis",
						TriggerOn: "mousemove",
						Enterable: false,
					}),
					charts.WithTitleOpts(opts.Title{
						Title:    graph_cache["name"].(string),
						Subtitle: "",
					}),
					charts.WithYAxisOpts(opts.YAxis{
						Name: "",
					}),
				)
				for x := 0; x < value_count; x++ {
					line.SetXAxis(graph_cache["times"].([]string)).
						AddSeries(fmt.Sprintf("value%d", x), graph_cache[fmt.Sprintf("value%d", x)].([]opts.LineData)).
						SetSeriesOptions(charts.WithLineChartOpts(opts.LineChart{Smooth: false}))
				}

				if err := line.Render(f); err != nil {
					log.Println(err)
				}

				log.Println("Chart created: lineChart.html")
			}
		}
	}
}

/**
* Count how many arrays in our JSON
*
 */
func array_count(jsonData []byte) (int, error) {
	var data map[string]interface{}

	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		return 0, err
	}

	count := 0

	for _, v := range data {
		if arr, ok := v.([]interface{}); ok {
			_ = arr
			count++
		}
	}

	return count, nil
}
