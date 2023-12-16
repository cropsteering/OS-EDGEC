package main

import (
	"log"
	"net/http"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
)

/**
* Setup HTTPD
*
 */
func Setup_Http() {
	log.Println("Starting HTTPD")
	http.HandleFunc("/", http_server)
	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func http_server(w http.ResponseWriter, r *http.Request) {
	log.Println("HTTP Request")
	MU.Lock()
	defer MU.Unlock()

	for k, v := range graph_cache {
		line := charts.NewLine()
		line.SetGlobalOptions(
			charts.WithInitializationOpts(opts.Initialization{Theme: types.ChartLine}),
			charts.WithTitleOpts(opts.Title{
				Title:    "Influx data",
				Subtitle: k,
			}))
			
		line.SetXAxis(time_cache).
			AddSeries(k, v).
			SetSeriesOptions(
				charts.WithLineChartOpts(opts.LineChart{
					ShowSymbol: false,
					Smooth:     false,
				}),
				charts.WithLabelOpts(opts.Label{
					Show: false,
				}),
			)
		line.Render(w)
	}
}
