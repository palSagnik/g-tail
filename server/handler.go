package server

import (
	"fmt"
	"logwatcher/broker"
	"logwatcher/watcher"
	"net/http"
)

// this will handle the server streaming
func Streamhandler(b *broker.Broker, logfile string, w http.ResponseWriter, r *http.Request) {

	// 1. Set SSE Headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// 2. Send the initial lines
	initialLines, _, err := watcher.GetLastNLines(logfile, 10)
	if err == nil {
		for _, line := range initialLines {
			fmt.Fprintf(w, "data: %s\n\n", line)
		}
		w.(http.Flusher).Flush()
	}

	// 3. Register New Client
	clientChan := make(chan string)
	b.NewClients <- clientChan

	// 4. Clean disconnection
	defer func() {
		b.Defunct <- clientChan
	}()

	// 5. Listening for updates
	ctx := r.Context()
	for {
		select {
		case msg := <-clientChan:
			fmt.Fprintf(w, "data: %s\n\n", msg)
			w.(http.Flusher).Flush()

		case <-ctx.Done():
			return 
		}
	}
}