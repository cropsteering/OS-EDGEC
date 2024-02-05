package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"slices"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var MQTT_CLIENT mqtt.Client
var MQTT_CONNECTED bool = false
var mqtt_topics []string
var Enable_Disco bool = false

func Setup_MQTT() {
	// Setup main MQTT connection
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tls://%s:%d", MQTT_BROKER, MQTT_PORT))
	tlsConfig := NewTlsConfig()
	opts.SetTLSConfig(tlsConfig)
	opts.SetClientID(MQTT_ID)
	opts.SetUsername(MQTT_USER)
	opts.SetPassword(MQTT_PASSWORD)
	opts.SetDefaultPublishHandler(messagepub_handler)
	opts.OnConnect = connect_handler
	opts.OnConnectionLost = connectlost_handler
	MQTT_CLIENT = mqtt.NewClient(opts)
	if token := MQTT_CLIENT.Connect(); token.Wait() && token.Error() != nil {
		R_LOG(token.Error().Error())
	}

	if _, err := os.Stat("topics.json"); os.IsNotExist(err) {
		mqtt_topics = append(mqtt_topics, MQTT_CONFIG)
		mqtt_topics = append(mqtt_topics, MQTT_STATUS)
		cerr := Cache_Array("topics.json", mqtt_topics)
		if cerr != nil {
			R_LOG("Error caching data: " + cerr.Error())
		}
	} else {
		topics, terr := Read_Array("topics.json")
		if terr != nil {
			R_LOG(terr.Error())
		} else {
			mqtt_topics = topics
		}
	}
	mqtt_sub(MQTT_CLIENT, MQTT_CONFIG)
	mqtt_sub(MQTT_CLIENT, MQTT_STATUS)
	mqtt_sub(MQTT_CLIENT, MQTT_USER+"/#")
}

/**
* Create TLS confid
*
 */
func NewTlsConfig() *tls.Config {
	certpool := x509.NewCertPool()
	ca, err := os.ReadFile("ca.pem")
	if err != nil {
		R_LOG(err.Error())
	}
	certpool.AppendCertsFromPEM(ca)
	return &tls.Config{
		RootCAs:            certpool,
		InsecureSkipVerify: true,
	}
}

/**
* Publish to MQTT
*
 */
func Mqtt_Publish(topic string, msg string) {
	if MQTT_CONNECTED {
		text := fmt.Sprint(msg)
		token := MQTT_CLIENT.Publish(topic, 0, false, text)
		token.Wait()
		R_LOG("MQTT publish: " + msg + "/" + topic)
		time.Sleep(time.Second)
	}
}

/**
* Subscribe to MQTT topics
*
 */
func mqtt_sub(client mqtt.Client, topic string) {
	if MQTT_CONNECTED {
		if token := client.Subscribe(topic, 0, nil); token.Wait() && token.Error() != nil {
			R_LOG("MQTT Token " + token.Error().Error())
		} else {
			R_LOG("Subscribed to " + topic)
		}
	}
}

/**
* MQTT Incoming message call back
*
 */
var messagepub_handler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	rec_msg := fmt.Sprintf("MQTT recieved: %s, topic: %s", msg.Payload(), msg.Topic())
	R_LOG(rec_msg)
	switch msg.Topic() {
	case MQTT_CONFIG:
		R_LOG("Config msg")
	case MQTT_STATUS:
		R_LOG("Status msg")
	default:
		Ingest_MQTT(Influx_Client, msg.Topic(), string(msg.Payload()))
	}
	MQTT_Disco(msg.Topic())
}

/**
* Discovery Mode
* Search for topics to auto generate values
*
 */
func MQTT_Disco(topic string) {
	mqtt_mu.Lock()
	if Enable_Disco {
		contains := slices.Contains(mqtt_topics, topic)
		if !contains {
			R_LOG("New topic: " + topic)
			mqtt_topics = append(mqtt_topics, topic)
			Append_String("topics.json", topic)
			Query_Values()
		}
	}
	mqtt_mu.Unlock()
}

/**
* MQTT Connect call back
*
 */
var connect_handler mqtt.OnConnectHandler = func(client mqtt.Client) {
	R_LOG("MQTT Connected")
	MQTT_CONNECTED = true
}

/**
* MQTT Disconnect call back
*
 */
var connectlost_handler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	R_LOG("MQTT Disconnected: " + err.Error())
	MQTT_CONNECTED = false
}
