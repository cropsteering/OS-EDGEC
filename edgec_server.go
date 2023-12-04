package main

import (
	"io"
	"log"
	"net/http"
	"time"
)

func main() {
	setup()
}

func setup() {
	go logic_loop()
	http.HandleFunc("/", http_server)
	err := http.ListenAndServe(":8081", nil)
	if (err != nil) {
		log.Fatal(err)
	}
}

func http_server(w http.ResponseWriter, r *http.Request) {
	log.Println("HTTP Request")
	io.WriteString(w, "Open Steering edge controller web portal")
}

func logic_loop() {
	for {
		log.Println("Logic tick")
		time.Sleep(time.Second)
	}
}
