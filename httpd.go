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
	http.HandleFunc("/", http_server)
	http.HandleFunc("/graphs", func(w http.ResponseWriter, r *http.Request) {
		Query_Topics()
		MU.Lock()
		http.ServeFile(w, r, "lineChart.html")
		MU.Unlock()
	})
	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func http_server(w http.ResponseWriter, r *http.Request) {
	log.Println("HTTP Request")
	http.ServeFile(w, r, "index.html")
}
