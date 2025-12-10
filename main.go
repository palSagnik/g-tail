package main

import (
	"fmt"
	"log"
	"logwatcher/broker"
	"logwatcher/server"
	"logwatcher/watcher"
	"net/http"
	"os"
)


const (
	LOGFILE = "./app.log"
	PORT = ":9001"
)

func main() {
	ensureLogFileExists(LOGFILE)

	// Broker
	bk := broker.NewBroker()
	go bk.Start()

	_, initialOffset, err := watcher.GetLastNLines(LOGFILE, 0)
	if err != nil {
		log.Println("error in intial logfile")
	}

	go watcher.Watch(LOGFILE, initialOffset, bk.Message)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "out.html")
	})
	
	http.HandleFunc("/log", func(w http.ResponseWriter, r *http.Request) {
		server.Streamhandler(bk, LOGFILE, w, r)
	})

	fmt.Printf("Log Watcher Server on port%s\n", PORT)

	log.Fatal(http.ListenAndServe(PORT, nil))
}

func ensureLogFileExists(logfile string) {
	f, _ := os.OpenFile(logfile, os.O_APPEND | os.O_CREATE | os.O_WRONLY, 0644)
	f.Close()
}