package main

/** MQTT Info */
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
var INFLUX_BUCKET = ""
var INFLUX_MEASUREMENT = ""

/** Debug Messages */
var DEBUG = true
