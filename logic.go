/**
* TODO: Add edit and cloning, weight, get powerc name automatically
*
 */

package main

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

var high_mu sync.Mutex
var pause_mu sync.Mutex
var lcache_mu sync.Mutex
var state_mu sync.Mutex

var logic_delay = 5
var logic_cache []map[string]interface{}
var state_cache = make(map[int]interface{})
var B_THEN bool = false
var then_start bool = false
var pause_logic []string
var then_cache []string
var uuid_cache string
var high_weight int = 0
var weight_list = make(map[string]int)

func Logic_Setup() {
	Load_Logic()

	ticker := time.NewTicker(time.Duration(logic_delay) * time.Second)
	go func() {
		for range ticker.C {
			Logic_Loop()
		}
	}()
}

func Reset_Logic() {
	lcache_mu.Lock()
	logic_cache = nil
	lcache_mu.Unlock()
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
			lcache_mu.Lock()
			logic_cache = append(logic_cache, temp_map)
			lcache_mu.Unlock()
			temp_map = nil
		}
	}
}

func Logic_Loop() {
	lcache_mu.Lock()
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
				weight_i, err3 := strconv.Atoi(slice[8])
				if err3 != nil {
					R_LOG(err3.Error())
				}
				run_logic(slice[0], slice[1], slice[2], reading_i, pin_i, slice[5], slice[6], slice[7], uuid, weight_i)
			}
		}
	}
	lcache_mu.Unlock()
}

func run_logic(sen_name string, val_name string, equ string, reading int, pin int, state string, powerc string, then string, uuid string, weight int) {
	flux_query := `
	from(bucket: "` + INFLUX_BUCKET + `")
	|> range(start: -1h)
	|> filter(fn: (r) => r["_measurement"] == "` + INFLUX_MEASUREMENT + `")
	|> filter(fn: (r) => r["topic"] == "` + sen_name + `")
	|> filter(fn: (r) => r["_field"] == "` + val_name + `")
	|> last()
	`

	if !Key_Exists_Int(uuid, weight_list) {
		weight_list[uuid] = weight
	}

	high_mu.Lock()
	if weight > high_weight {
		high_weight = weight
	}
	high_mu.Unlock()

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
					pause_mu.Lock()
					if !String_Exists(uuid, pause_logic) {
						high_mu.Lock()
						if weight >= high_weight {
							if run_equations(equ, db_fvalue, f_reading, sen_name, state, pin, powerc) {
								if B_THEN {
									pause_logic = append(pause_logic, uuid)
									then_start = true
									uuid_cache = uuid
								} else {
									high_weight = 0
								}
							}
						}
						high_mu.Unlock()
					}
					pause_mu.Unlock()
				}
			}
		}
	} else {
		if !String_Exists(uuid_cache, then_cache) {
			then_cache = append(then_cache, uuid_cache)
			go then_logic(uuid_cache, flux_query, sen_name, equ, reading, pin, state, powerc, weight)
		}
		then_start = false
	}
}

func then_logic(temp_uuid string, flux_query string, sen_name string, equ string, reading int, pin int, state string, powerc string, weight int) {
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
						pause_mu.Lock()
						pause_logic = String_Delete(temp_uuid, pause_logic)
						pause_mu.Unlock()
						then_cache = String_Delete(temp_uuid, then_cache)
						high_mu.Lock()
						if weight >= high_weight {
							high_weight = 0
						}
						high_mu.Unlock()
						ticker.Stop()
					}
				}
			}
		}
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
	state_mu.Lock()
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
	state_mu.Unlock()
}
