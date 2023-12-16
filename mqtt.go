package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var MQTT_CLIENT mqtt.Client
var MQTT_CONNECTED bool = false

func Setup_MQTT() {
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
		log.Println(token.Error())
	}

	// Sub to MQTT topics
	topics, err := Read_Array("topics.json")
	if err != nil {
		log.Println("Error reading JSON ", err)
	} else {
		topics = append(topics, MQTT_CONFIG)
		topics = append(topics, MQTT_STATUS)
		mqtt_sub(MQTT_CLIENT, topics)
	}

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

func mqtt_sub(client mqtt.Client, topics []string) {
	if MQTT_CONNECTED {
		for _, topic := range topics {
			// Subscribe to each topic
			if token := client.Subscribe(topic, 0, nil); token.Wait() && token.Error() != nil {
				log.Println(token.Error())
			} else {
				log.Printf("Subscribed to %s", topic)
			}
		}
	}
}

/**
* Publish to MQTT
*
 */
func mqtt_publish(topic string, msg string) {
	if MQTT_CONNECTED {
		text := fmt.Sprint(msg)
		token := MQTT_CLIENT.Publish(topic, 0, false, text)
		token.Wait()
		log.Printf("MQTT publish: %s to %s", msg, topic)
		time.Sleep(time.Second)
	}
}

/**
* MQTT Incoming message call back
*
 */
var messagepub_handler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	log.Printf("MQTT recieved: %s, topic: %s\n", msg.Payload(), msg.Topic())
	switch msg.Topic() {
	case MQTT_CONFIG:
		log.Println("Config msg")
	case MQTT_STATUS:
		log.Println("Status msg")
	default:
		Ingest_MQTT(Influx_Client, msg)
	}
}

/**
* MQTT Connect call back
*
 */
var connect_handler mqtt.OnConnectHandler = func(client mqtt.Client) {
	log.Println("MQTT Connected")
	MQTT_CONNECTED = true
}

/**
* MQTT Disconnect call back
*
 */
var connectlost_handler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	log.Printf("MQTT Disconnected: %v", err)
	MQTT_CONNECTED = false
}
