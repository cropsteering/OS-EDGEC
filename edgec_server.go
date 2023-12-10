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
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

/** MQTT Info */
var MQTT_CLIENT mqtt.Client
var MQTT_BROKER = ""
var MQTT_ID = ""
var MQTT_USER = ""
var MQTT_PASSWORD = ""
var MQTT_CONFIG = MQTT_USER + "/" + MQTT_ID + "/config"
var MQTT_STATUS = MQTT_USER + "/" + MQTT_ID + "/sataus"
var MQTT_PORT = 8883

/**
* Programs main call
*
 */
func main() {
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
	go start_http()

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
* Start HTTPD
*
 */
func start_http() {
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
	io.WriteString(w, "Open Steering edge controller web portal")
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
