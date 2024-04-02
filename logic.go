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
var scache_mu sync.Mutex
var state_mu sync.Mutex

var logic_delay = 5
var logic_cache []map[string]interface{}
var sched_cache []map[string]interface{}
var state_cache = make(map[int]interface{})
var B_THEN bool = false
var then_start bool = false
var pause_logic []string
var then_cache []string
var uuid_cache string
var high_weight int = 0
var sched_events []Sched_Event
var cancel_events = make(chan struct{})

type Sched_Event struct {
	UUID         string
	DayOfWeek    time.Weekday
	Time         time.Time
	Pin          int
	P_state      string
	P_controller string
	Zone         string
}

func Logic_Setup() {
	Load_Logic()
	Load_Sched()

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

func Reset_Sched() {
	close(cancel_events)
	scache_mu.Lock()
	sched_cache = nil
	sched_events = nil
	scache_mu.Unlock()
}

func Load_Logic() {
	logic_json, err := Read_Map("logic.json")
	if err != nil {
		R_LOG(err.Error())
	} else {
		v_keys, v_values := Iterate_Map(logic_json)
		for v_index, v_name := range v_keys {
			temp_map := make(map[string]interface{})
			temp_value, _ := Iterate_Interface(v_values[v_index])
			temp_map[v_name] = temp_value
			lcache_mu.Lock()
			logic_cache = append(logic_cache, temp_map)
			lcache_mu.Unlock()
			temp_map = nil
		}
	}
}

func Load_Sched() {
	sched_json, err := Read_Map("sched.json")
	if err != nil {
		R_LOG(err.Error())
	} else {
		v_keys, v_values := Iterate_Map(sched_json)
		for v_index, v_name := range v_keys {
			temp_map := make(map[string]interface{})
			temp_value, _ := Iterate_Interface(v_values[v_index])
			temp_map[v_name] = temp_value
			scache_mu.Lock()
			sched_cache = append(sched_cache, temp_map)
			scache_mu.Unlock()
			temp_map = nil
		}
	}

	scache_mu.Lock()
	for _, v := range sched_cache {
		for uuid, values := range v {
			if slice, ok := values.([]string); ok {
				day := slice[0]
				time_stamp := slice[1]
				pin_i, err := strconv.Atoi(slice[2])
				if err != nil {
					R_LOG(err.Error())
				}
				pin_state := slice[3]
				p_controller := slice[4]
				zone := slice[5]
				setup_sched(uuid, day, time_stamp, pin_i, pin_state, p_controller, zone)
			}
		}
	}
	scache_mu.Unlock()

	cancel_events = make(chan struct{})
	for _, event := range sched_events {
		go func(event Sched_Event) {
			run_scheduler(event, cancel_events)
		}(event)
	}
}

func run_scheduler(event Sched_Event, cancel <-chan struct{}) {
	for {
		select {
		case <-cancel:
			return
		default:
			now := time.Now()
			sleep_time := 30 * time.Second
			if now.Weekday() == event.DayOfWeek && now.Hour() == event.Time.Hour() && now.Minute() == event.Time.Minute() {
				R_LOG("Scheduled event triggered")
				high_mu.Lock()
				previous_weight := high_weight
				if 999 > high_weight {
					high_weight = 999
					pin_switch_sched(event.Zone, event.P_state, event.Pin, event.P_controller)
					high_weight = previous_weight
					sleep_time = 1 * time.Minute
				}
				high_mu.Unlock()
			}
			time.Sleep(sleep_time)
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

func setup_sched(uuid string, day string, time_stamp string, pin int, pin_state string, p_controller string, zone string) {
	if !hasUUID(uuid) {
		week_day, _ := StringToWeekday(day)
		event_time, _ := time.Parse("1504", time_stamp)
		new_event := Sched_Event{
			UUID:         uuid,
			DayOfWeek:    week_day,
			Time:         event_time,
			Pin:          pin,
			P_state:      pin_state,
			P_controller: p_controller,
			Zone:         zone,
		}
		sched_events = append(sched_events, new_event)
	}
}

func hasUUID(uuidToFind string) bool {
	for _, event := range sched_events {
		if event.UUID == uuidToFind {
			return true
		}
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

func pin_switch_sched(zone string, state string, pin int, powerc string) {
	state_mu.Lock()
	if value, exists := state_cache[pin]; exists && value == state {
		// Ignore
	} else {
		switch state {
		case "on":
			pin_cmd := "0+" + strconv.Itoa(pin)
			power_control := MQTT_USER + "/" + zone + "/" + powerc + "/control"
			Mqtt_Publish(power_control, pin_cmd)
			state_cache[pin] = state
		case "off":
			pin_cmd := "1+" + strconv.Itoa(pin)
			power_control := MQTT_USER + "/" + zone + "/" + powerc + "/control"
			Mqtt_Publish(power_control, pin_cmd)
			state_cache[pin] = state
		}
	}
	state_mu.Unlock()
}
