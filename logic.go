/**
* TODO: Add edit and cloning, weight, get powerc name automatically
*
 */

package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

/**
* Basic logic loop
* TODO: Grab values from values.json
*
 */
var logic_delay = 5
var logic_cache []map[string]interface{}

func Logic_Setup() {
	Load_Logic()

	ticker := time.NewTicker(time.Duration(logic_delay) * time.Second)
	go func() {
		for range ticker.C {
			Logic_Loop()
		}
	}()
}

func Load_Logic() {
	logic_json, err := Read_Map("logic.json")
	if err != nil {
		log.Println(err)
	} else {
		v_keys, v_values := Iterate_Map(logic_json)
		for v_index, v_name := range v_keys {
			temp_map := make(map[string]interface{})
			temp_value := Iterate_Interface(v_values[v_index])
			temp_map[v_name] = temp_value
			logic_cache = append(logic_cache, temp_map)
			temp_map = nil
		}
	}
}

func Logic_Loop() {
	for _, v := range logic_cache {
		for _, values := range v {
			if slice, ok := values.([]string); ok {
				reading_i, err := strconv.Atoi(slice[3])
				if err != nil {
					log.Println(err)
				}
				pin_i, err2 := strconv.Atoi(slice[4])
				if err2 != nil {
					log.Println(err2)
				}
				run_logic(slice[0], slice[1], slice[2], reading_i, pin_i, slice[5], slice[6])
			}
		}
	}
}

func run_logic(sen_name string, val_name string, equ string, reading int, pin int, state string, powerc string) {
	flux_query := `
	from(bucket: "` + INFLUX_BUCKET + `")
	|> range(start: -1h)
	|> filter(fn: (r) => r["_measurement"] == "` + INFLUX_MEASUREMENT + `")
	|> filter(fn: (r) => r["topic"] == "` + sen_name + `")
	|> filter(fn: (r) => r["_field"] == "` + val_name + `")
	|> last()
	`
	results, err := Query_DB(flux_query)
	if err != nil {
		log.Println(err)
	} else {
		for results.Next() {
			db_value := fmt.Sprintf("%f", results.Record().Value())
			db_fvalue, err := strconv.ParseFloat(db_value, 64)
			f_reading := float64(reading)
			if err != nil {
				log.Println(err)
			} else {
				switch equ {
				case "equa":
					if db_fvalue == f_reading {
						pin_switch(sen_name, state, pin, powerc)
					}
				case "nequa":
					if db_fvalue != f_reading {
						pin_switch(sen_name, state, pin, powerc)
					}
				case "less":
					if db_fvalue < f_reading {
						pin_switch(sen_name, state, pin, powerc)
					}
				case "great":
					if db_fvalue > f_reading {
						pin_switch(sen_name, state, pin, powerc)
					}
				}
			}
		}
	}
}

func pin_switch(sen_name string, state string, pin int, powerc string) {
	parts := strings.Split(sen_name, "/")
	switch state {
	case "on":
		pin_cmd := "0+" + strconv.Itoa(pin)
		power_control := MQTT_USER + "/" + parts[1] + "/" + powerc + "/control"
		Mqtt_Publish(power_control, pin_cmd)
		log.Println("pin " + strconv.Itoa(pin) + " on")
	case "off":
		pin_cmd := "1+" + strconv.Itoa(pin)
		power_control := MQTT_USER + "/" + parts[1] + "/" + powerc + "/control"
		Mqtt_Publish(power_control, pin_cmd)
		log.Println("pin " + strconv.Itoa(pin) + " off")
	}
}
