package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"time"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

var Influx_Client influxdb2.Client
var graph_cache = make(map[string]interface{})

func Setup_Influxdb() {
	R_LOG("Starting InfluxDB")
	Influx_Client = influxdb2.NewClient(INFLUX_URL, INFLUX_TOKEN)
	defer Influx_Client.Close()

	// Ping InfluxDB server to check status
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := Influx_Client.Ping(ctx)
	if err != nil {
		R_LOG("InfluxDB not connected " + err.Error())
	} else {
		R_LOG("InfluxDB connected")
		Query_Topics()
	}
}

/**
* Query MQTT topics in Influx
* Then generate chart data, save to html
*
 */
func Query_Topics() {
	graph_mu.Lock()
	defer graph_mu.Unlock()
	cached_array, err := Read_Array("topics.json")
	if err != nil {
		R_LOG("Error reading array from file: " + err.Error())
	} else {
		f, err := os.Create("lineChart.html")
		if err != nil {
			R_LOG(err.Error())
		} else {
			f.WriteString("<center>Openly Automated</center>")
			R_LOG("Chart created: lineChart.html")
			for _, v := range cached_array {
				if v != MQTT_STATUS && v != MQTT_CONFIG {
					graph_cache = make(map[string]interface{})
					flux_query := `
								from(bucket: "` + INFLUX_BUCKET + `")
								|> range(start: -1h)
								|> filter(fn: (r) => r["_measurement"] == "` + INFLUX_MEASUREMENT + `")
								|> filter(fn: (r) => r["topic"] == "` + v + `")
								|> filter(fn: (r) => r._field != "state")
								|> yield(name: "mean")
								`
					results, err := Query_DB(flux_query)
					if err != nil {
						R_LOG(err.Error())
					} else {
						var influx_data []opts.LineData
						var time_data []string
						var value_names []string
						var last_value string
						var v_name string

						for results.Next() {
							v_name = results.Record().Field()
							contains := slices.Contains(value_names, v_name)
							if !contains {
								value_names = append(value_names, v_name)
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
							R_LOG("Query processing error: " + results.Err().Error())
						}

						if graph_cache["name"] != nil {
							json_obj, _ := json.MarshalIndent(graph_cache, "", "  ")
							value_count, _ := Array_Count(json_obj)
							// -1 because of []times
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
								R_LOG(err.Error())
							}
						} else {
							f.WriteString("Error loading graph data.")
						}
					}
					graph_cache = nil
				}
			}
		}
		f.WriteString("<br /><a href='index.html'><button>Back to dashboard</button></a>")
	}
}

func Query_Values() {
	influx_mu.Lock()
	defer influx_mu.Unlock()
	cached_array, err := Read_Array("topics.json")
	if err != nil {
		R_LOG("Error reading array: " + err.Error())
	} else {
		var topics_cache = make(map[string]interface{})
		is_empty := true
		for _, v := range cached_array {
			if v != MQTT_STATUS && v != MQTT_CONFIG {
				var value_names []string
				flux_query := `
				from(bucket: "` + INFLUX_BUCKET + `")
				|> range(start: -1h)
				|> filter(fn: (r) => r["_measurement"] == "` + INFLUX_MEASUREMENT + `")
				|> filter(fn: (r) => r["topic"] == "` + v + `")
				|> yield(name: "mean")
				`
				results, err := Query_DB(flux_query)
				if err != nil {
					R_LOG("Query: " + err.Error())
				} else {
					var v_name string
					for results.Next() {
						v_name = results.Record().Field()
						contains := slices.Contains(value_names, v_name)
						if !contains {
							value_names = append(value_names, v_name)
						}
						is_empty = false
					}
				}
				topics_cache[v] = value_names
				value_names = nil
			}
		}
		if is_empty {
			R_LOG("No values found, trying again in 1 minute.")
			go func() {
				R_LOG("Querying values")
				time.Sleep(1 * time.Minute)
				Query_Values()
			}()
		} else {
			cache_err := Cache_Map(topics_cache, "values.json")
			if cache_err != nil {
				R_LOG("Error caching array: " + cache_err.Error())
			} else {
				R_LOG("Loaded topics cache")
			}
		}
	}
}

func Query_DB(query string) (*api.QueryTableResult, error) {
	queryAPI := Influx_Client.QueryAPI(INFLUX_ORG)
	result, err := queryAPI.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}

	return result, nil
}
