package main

import "time"

/**
* Basic logic loop
* # WARNING # This is just for testing purposes
*
 */
var flip = false

func Logic_Loop() {
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
