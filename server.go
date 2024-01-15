package main

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var graph_mu sync.Mutex
var logic_mu sync.Mutex
var mqtt_mu sync.Mutex
var influx_mu sync.Mutex

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
