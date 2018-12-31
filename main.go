package main

import (
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/gorilla/mux"
	"github.com/krantius/raft"
)

func main() {
	id := os.Getenv("NODE_ID")
	if id == "" {
		panic("NODE_ID not set")
	}

	peers := os.Getenv("NODE_PEERS")

	port := 8001

	portArgs := os.Getenv("NODE_PORT")
	if portArgs != "" {
		tport, err := strconv.Atoi(portArgs)
		if err == nil {
			port = tport
		}
	}

	raftNode := raft.NewNode(id, port, strings.Split(peers, ","))
	go raftNode.Do()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	<-c

	r := mux.NewRouter()
	sr := r.PathPrefix("/api").Subrouter()
	sr.Path("/status").Methods("GET").HandlerFunc(node.Status)

	go http.ListenAndServe(":8000", r)
}
