/**
 * @file edgec_server.go
 * @author Jamie Howse (r4wknet@gmail.com)
 * @brief
 * @version 0.1
 * @date 2023-06-10
 *
 * @copyright Copyright (c) 2023
 *
 */

package main

/** Imports */
import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

/** MQTT Info */
var MQTT_CLIENT mqtt.Client
var MQTT_BROKER = ""
var MQTT_ID = "EdgeC"
var MQTT_USER = ""
var MQTT_PASSWORD = ""
var MQTT_CONFIG = MQTT_USER + "/" + MQTT_ID + "/config"
var MQTT_STATUS = MQTT_USER + "/" + MQTT_ID + "/sataus"
var MQTT_PORT = 8883

/** InfluxDB Info */
var INFLUX_URL = ""
var INFLUX_TOKEN = ""
var INFLUX_ORG = ""

/** Channels */
var line_chan = make(chan []opts.LineData)
var time_chan = make(chan []opts.LineData)

/** Cache */
var line_cache []opts.LineData
var time_cache []opts.LineData

/**
* Programs main call
*
 */
func main() {
	log.Println("Starting Edge Controller")
	setup()
}

/**
* Main setup
*
 */
func setup() {
	/** Create chan for signal support */
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	/** Start any Go routines here */
	go logic_loop()
	go setup_http()
	go setup_influxdb()

	/** Setup MQTT server on main routine */
	mqtt_setup()

	<-done
}

/**
* Basic logic loop
* # WARNING # This is just for testing purposes
*
 */
var flip = false

func logic_loop() {
	for {
		time.Sleep(time.Second)
		if flip {
			mqtt_publish("r4wk/Zone1/powerc/control", "0+1")
			flip = false
		} else {
			mqtt_publish("r4wk/Zone1/powerc/control", "1+1")
			flip = true
		}
	}
}

/**
* Setup HTTPD
*
 */
func setup_http() {
	log.Println("Starting HTTPD")
	http.HandleFunc("/", http_server)
	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		log.Fatal(err)
	}
}

/**
* Serve HTTP
*
 */
func http_server(w http.ResponseWriter, r *http.Request) {
	log.Println("HTTP Request")
	if len(line_cache) <= 0 {
		line_cache = <-line_chan
	}
	if len(time_cache) <= 0 {
		time_cache = <-time_chan
	}
	line := charts.NewLine()
	line.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{Theme: types.ChartLine}),
		charts.WithTitleOpts(opts.Title{
			Title:    "Influx data",
			Subtitle: "Sensors",
		}))

	line.SetXAxis(time_cache).
		AddSeries("Category A", line_cache).
		SetSeriesOptions(
			charts.WithLineChartOpts(opts.LineChart{
				ShowSymbol: true,
				Smooth:     false,
			}),
			charts.WithLabelOpts(opts.Label{
				Show: true,
			}),
		)
	line.Render(w)
}

/**
* Setup InfluxDB
*
 */
func setup_influxdb() {
	client := influxdb2.NewClient(INFLUX_URL, INFLUX_TOKEN)
	defer client.Close()
	queryAPI := client.QueryAPI(INFLUX_ORG)
	query := `from(bucket: "opensteering")
                |> range(start: -1m)
                |> filter(fn: (r) => r._measurement == "mqtt_consumer")
			`
	results, err := queryAPI.Query(context.Background(), query)
	if err != nil {
		log.Println(err)
	} else {
		var influx_data []string
		var time_data []string
		for results.Next() {
			influx_value := fmt.Sprintf("%v", results.Record().Value())
			//mqtt_topic := fmt.Sprintf("%v", results.Record().ValueByKey("topic"))
			mqtt_time := fmt.Sprintf("%v", results.Record().ValueByKey("_time"))
			time_data = append(time_data, mqtt_time)
			influx_data = append(influx_data, influx_value)
		}
		line_cache = nil
		time_cache = nil
		line_chan <- get_line(influx_data)
		time_chan <- get_time(time_data)
		if err := results.Err(); err != nil {
			log.Println(err)
		}
	}
}

func get_line(idata []string) []opts.LineData {
	ldata := make([]opts.LineData, 0)
	size := len(idata)
	for x := 0; x < size; x++ {
		ldata = append(ldata, opts.LineData{Name: fmt.Sprintf("Value %d", x+1), Value: idata[x]})
	}
	return ldata
}

func get_time(tdata []string) []opts.LineData {
	ldata := make([]opts.LineData, 0)
	size := len(tdata)
	for x := 0; x < size; x++ {
		ldata = append(ldata, opts.LineData{Value: tdata[x]})
	}
	return ldata
}

/**
* Setup MQTT
*
 */
func mqtt_setup() {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tls://%s:%d", MQTT_BROKER, MQTT_PORT))
	tlsConfig := NewTlsConfig()
	opts.SetTLSConfig(tlsConfig)
	opts.SetClientID(MQTT_ID)
	opts.SetUsername(MQTT_USER)
	opts.SetPassword(MQTT_PASSWORD)
	opts.SetDefaultPublishHandler(messagePubHandler)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	MQTT_CLIENT = mqtt.NewClient(opts)
	if token := MQTT_CLIENT.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	/** Sub to MQTT */
	mqtt_sub()
}

/**
* Publish to MQTT
*
 */
func mqtt_publish(topic string, msg string) {
	text := fmt.Sprint(msg)
	token := MQTT_CLIENT.Publish(topic, 0, false, text)
	token.Wait()
	log.Printf("MQTT publish: %s to %s", msg, topic)
	time.Sleep(time.Second)
}

/**
* Subscribe to MQTT
*
 */
func mqtt_sub() {
	token := MQTT_CLIENT.Subscribe(MQTT_CONFIG, 1, nil)
	token.Wait()
	log.Printf("MQTT subscribed to topic: %s", MQTT_CONFIG)
	token = MQTT_CLIENT.Subscribe(MQTT_STATUS, 1, nil)
	token.Wait()
	log.Printf("MQTT subscribed to topic: %s", MQTT_STATUS)
}

/**
* Create TLS confid
*
 */
func NewTlsConfig() *tls.Config {
	certpool := x509.NewCertPool()
	ca, err := os.ReadFile("ca.pem")
	if err != nil {
		log.Fatalln(err.Error())
	}
	certpool.AppendCertsFromPEM(ca)
	return &tls.Config{
		RootCAs:            certpool,
		InsecureSkipVerify: true,
	}
}

/**
* MQTT Incoming message call back
*
 */
var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	log.Printf("MQTT recieved: %s, topic: %s\n", msg.Payload(), msg.Topic())
}

/**
* MQTT Connect call back
*
 */
var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	log.Println("MQTT Connected")
}

/**
* MQTT Disconnect call back
*
 */
var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	log.Printf("MQTT Disconnected: %v", err)
}
