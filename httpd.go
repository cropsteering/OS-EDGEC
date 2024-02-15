package main

import (
	"html/template"
	"net/http"
	"strconv"

	"github.com/google/uuid"
)

/**
* Setup HTTPD
*
 */
func Setup_Http() {
	R_LOG("Starting HTTPD")
	http.HandleFunc("/", index_server)
	http.HandleFunc("/graphs", graph_server)
	http.HandleFunc("/logic", logic_server)
	http.HandleFunc("/disco", disco_server)
	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		R_LOG(err.Error())
	}
}

func index_server(w http.ResponseWriter, r *http.Request) {
	R_LOG("HTTP Request, index")
	http.ServeFile(w, r, "index.html")
}

func graph_server(w http.ResponseWriter, r *http.Request) {
	R_LOG("HTTP Request, graphs")
	Query_Topics()
	graph_mu.Lock()
	http.ServeFile(w, r, "lineChart.html")
	graph_mu.Unlock()
}

func logic_server(w http.ResponseWriter, r *http.Request) {
	R_LOG("HTTP Request, logic")
	logic_mu.Lock()
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusInternalServerError)
		return
	}

	button_name := r.FormValue("submit")

	switch button_name {
	case "Add":
		R_LOG("Adding logic")
		add_logic(r)
	case "Delete":
		R_LOG("Deleting logic")
		del_logic(r)
	default:
		// Ignore
	}

	Query_Values()
	tmpl, err := template.New("logic").Parse(Build_Logic())
	if err != nil {
		R_LOG(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		err = tmpl.ExecuteTemplate(w, "logic", "")
		if err != nil {
			R_LOG(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
	logic_mu.Unlock()
}

func del_logic(r *http.Request) {
	logic_json, err := Read_Map("logic.json")
	if err != nil {
		R_LOG(err.Error())
	} else {
		delete(logic_json, r.FormValue("uuid"))
		temp_err := Cache_Map(logic_json, "logic.json")
		if temp_err != nil {
			R_LOG(temp_err.Error())
		}
		Reset_Logic()
		Load_Logic()
	}
}

func add_logic(r *http.Request) {
	if r.Method == http.MethodPost {
		UUID := uuid.New()
		var temp []string
		var temp_map = make(map[string]interface{})
		temp = append(temp, r.FormValue("SENSOR"))
		temp = append(temp, r.FormValue("READING"))
		temp = append(temp, r.FormValue("EQU"))
		temp = append(temp, r.FormValue("VALUE"))
		temp = append(temp, r.FormValue("PIN"))
		temp = append(temp, r.FormValue("STATUS"))
		temp = append(temp, r.FormValue("POWERC"))
		if Is_StringEmpty(r.FormValue("THEN")) {
			temp = append(temp, "FALSE")
		} else {
			temp = append(temp, r.FormValue("THEN"))
		}
		temp = append(temp, r.FormValue("WEIGHT"))
		temp_map[UUID.String()] = temp
		temp_err := Append_Map(temp_map, "logic.json")
		if temp_err != nil {
			R_LOG(temp_err.Error())
		}
		temp = nil
		temp_map = nil
		Reset_Logic()
		Load_Logic()
	}
}

func disco_server(w http.ResponseWriter, r *http.Request) {
	R_LOG("HTTP Request, disco")
	mqtt_mu.Lock()
	if r.Method == http.MethodPost {
		switch r.FormValue("Disco") {
		case "Enable":
			R_LOG("Discovery mode enabled")
			Enable_Disco = true
		case "Disable":
			R_LOG("Discovery mode disabled")
			Enable_Disco = false
		}
	}
	tmpl, err := template.New("disco").Parse(Disco_HTML)
	if err != nil {
		R_LOG(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		err = tmpl.ExecuteTemplate(w, "disco", strconv.FormatBool(Enable_Disco))
		if err != nil {
			R_LOG(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
	mqtt_mu.Unlock()
}
