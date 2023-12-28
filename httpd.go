package main

import (
	"log"
	"net/http"
)

/**
* Setup HTTPD
*
 */
func Setup_Http() {
	log.Println("Starting HTTPD")
	http.HandleFunc("/", index_server)
	http.HandleFunc("/graphs", graph_server)
	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func index_server(w http.ResponseWriter, r *http.Request) {
	log.Println("HTTP Request, index")
	http.ServeFile(w, r, "index.html")
}

func graph_server(w http.ResponseWriter, r *http.Request) {
	log.Println("HTTP Request, graphs")
	Query_Topics()
	MU.Lock()
	http.ServeFile(w, r, "lineChart.html")
	MU.Unlock()
}
