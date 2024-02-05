/**
* TODO: Add edit and cloning, weight, get powerc name automatically,
* if state already exists, dont do anything
*
 */

package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

var logic_delay = 5
var logic_cache []map[string]interface{}
var state_cache = make(map[int]interface{})
var B_THEN bool = false
var then_start bool = false
var pause_logic []string
var then_cache = make(map[string]string)
var uuid_cache string

func Logic_Setup() {
	Load_Logic()
	Query_Values()

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
		R_LOG(err.Error())
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
		for uuid, values := range v {
			if slice, ok := values.([]string); ok {
				reading_i, err := strconv.Atoi(slice[3])
				if err != nil {
					R_LOG(err.Error())
				}
				pin_i, err2 := strconv.Atoi(slice[4])
				if err2 != nil {
					R_LOG(err2.Error())
				}
				run_logic(slice[0], slice[1], slice[2], reading_i, pin_i, slice[5], slice[6], slice[7], uuid)
			}
		}
	}
}

func run_logic(sen_name string, val_name string, equ string, reading int, pin int, state string, powerc string, then string, uuid string) {
	flux_query := `
	from(bucket: "` + INFLUX_BUCKET + `")
	|> range(start: -1h)
	|> filter(fn: (r) => r["_measurement"] == "` + INFLUX_MEASUREMENT + `")
	|> filter(fn: (r) => r["topic"] == "` + sen_name + `")
	|> filter(fn: (r) => r["_field"] == "` + val_name + `")
	|> last()
	`

	if then == "TRUE" {
		B_THEN = true
	} else {
		B_THEN = false
	}

	if !then_start {
		results, err := Query_DB(flux_query)
		if err != nil {
			R_LOG(err.Error())
		} else {
			for results.Next() {
				db_value := fmt.Sprintf("%f", results.Record().Value())
				db_fvalue, err := strconv.ParseFloat(db_value, 64)
				f_reading := float64(reading)
				if err != nil {
					R_LOG(err.Error())
				} else {
					if !String_Exists(uuid, pause_logic) {
						if run_equations(equ, db_fvalue, f_reading, sen_name, state, pin, powerc) {
							if B_THEN {
								pause_logic = append(pause_logic, uuid)
								then_start = true
								uuid_cache = uuid
							}
						}
					}
				}
			}
		}
	} else {
		if !Key_Exists(uuid, then_cache) {
			then_cache[uuid] = uuid_cache
			go func() {
				ticker := time.NewTicker(time.Duration(logic_delay) * time.Second)
				for range ticker.C {
					results, err := Query_DB(flux_query)
					if err != nil {
						R_LOG(err.Error())
					} else {
						for results.Next() {
							db_value := fmt.Sprintf("%f", results.Record().Value())
							db_fvalue, err := strconv.ParseFloat(db_value, 64)
							f_reading := float64(reading)
							if err != nil {
								R_LOG(err.Error())
							} else {
								if run_equations(equ, db_fvalue, f_reading, sen_name, state, pin, powerc) {
									delete(then_cache, uuid)
									pause_logic = String_Delete(then_cache[uuid], pause_logic)
									then_start = false
									ticker.Stop()
								}
							}
						}
					}
				}
			}()
		}
		then_start = false
	}
}

func run_equations(equ string, db_fvalue float64, f_reading float64, sen_name string, state string, pin int, powerc string) bool {
	switch equ {
	case "equa":
		if db_fvalue == f_reading {
			pin_switch(sen_name, state, pin, powerc)
			return true
		}
	case "nequa":
		if db_fvalue != f_reading {
			pin_switch(sen_name, state, pin, powerc)
			return true
		}
	case "less":
		if db_fvalue < f_reading {
			pin_switch(sen_name, state, pin, powerc)
			return true
		}
	case "lessequ":
		if db_fvalue <= f_reading {
			pin_switch(sen_name, state, pin, powerc)
			return true
		}
	case "great":
		if db_fvalue > f_reading {
			pin_switch(sen_name, state, pin, powerc)
			return true
		}
	case "greatequ":
		if db_fvalue >= f_reading {
			pin_switch(sen_name, state, pin, powerc)
			return true
		}
	default:
		return false
	}
	return false
}

func pin_switch(sen_name string, state string, pin int, powerc string) {
	parts := strings.Split(sen_name, "/")
	if value, exists := state_cache[pin]; exists && value == state {
		// Ignore
	} else {
		switch state {
		case "on":
			pin_cmd := "0+" + strconv.Itoa(pin)
			power_control := MQTT_USER + "/" + parts[1] + "/" + powerc + "/control"
			Mqtt_Publish(power_control, pin_cmd)
			state_cache[pin] = state
		case "off":
			pin_cmd := "1+" + strconv.Itoa(pin)
			power_control := MQTT_USER + "/" + parts[1] + "/" + powerc + "/control"
			Mqtt_Publish(power_control, pin_cmd)
			state_cache[pin] = state
		}
	}
}
