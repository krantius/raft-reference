package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/gorilla/mux"
	"github.com/krantius/logging"
	"github.com/krantius/raft"
	"github.com/krantius/raft-reference/store"
)

type putArgs struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type server struct {
	rft  *raft.Raft
	data store.Store
}

func main() {
	setLogLevel()

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

	httpPort := 8000

	httpPortArgs := os.Getenv("HTTP_PORT")
	if httpPortArgs != "" {
		tport, err := strconv.Atoi(httpPortArgs)
		if err == nil {
			httpPort = tport
		}
	}

	cfg := raft.Config{
		ID:    id,
		Port:  port,
		Peers: strings.Split(peers, ","),
	}

	inMemStore := store.New()

	ctx, cancel := context.WithCancel(context.Background())

	s := &server{
		rft:  raft.New(ctx, cfg, inMemStore),
		data: inMemStore,
	}

	go s.rft.Start()

	r := mux.NewRouter()
	r.Path("/kv").Methods("GET").HandlerFunc(s.status)
	r.Path("/kv").Methods("PUT").HandlerFunc(s.put)
	r.Path("/status").Methods("GET").HandlerFunc(s.raftDump)

	go http.ListenAndServe(fmt.Sprintf(":%d", httpPort), r)

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	<-c
	cancel()
}

func (s *server) put(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	args := &putArgs{}
	if err := json.Unmarshal(b, args); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	c := raft.Command{
		Op:  raft.Set,
		Key: args.Key,
		Val: []byte(args.Value),
	}

	logging.Infof("HTTP Applying %s %s", args.Key, args.Value)

	err = s.rft.Apply(c)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Write([]byte("Success"))
}

func (s *server) status(w http.ResponseWriter, r *http.Request) {
	w.Write(s.data.Dump())
}

func (s *server) raftDump(w http.ResponseWriter, r *http.Request) {
	w.Write(s.rft.Dump())
	w.Write([]byte("\n"))
	w.Write(s.data.Dump())
}

func setLogLevel() {
	level := os.Getenv("LOG_LEVEL")
	if level != "" {
		switch level {
		case "trace":
			logging.SetLevel(logging.TRACE)
		case "debug":
			logging.SetLevel(logging.DEBUG)
		case "info":
			logging.SetLevel(logging.INFO)
		case "warning":
			logging.SetLevel(logging.WARNING)
		case "error":
			logging.SetLevel(logging.ERROR)
		}
	}
}
