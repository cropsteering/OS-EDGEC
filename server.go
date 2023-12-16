package main

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/go-echarts/go-echarts/v2/opts"
)

var MU sync.Mutex

// Cache
var graph_cache = make(map[string][]opts.LineData)
var time_cache []string

func main() {
	/** Create chan for signal support (CTRL+C) */
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go Setup_Http()
	go Setup_Influxdb()
	go Setup_MQTT()
	go Logic_Loop()

	<-done
}
