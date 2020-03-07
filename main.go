package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
)

// Command line parameters
var inmem bool
var nodeID int

func init() {
	flag.BoolVar(&inmem, "inmem", false, "Use in-memory storage for Raft")
	flag.IntVar(&nodeID, "id", -1, "Node ID")
}

func main() {
	flag.Parse()

	if flag.NArg() == 0 {
		fmt.Fprintf(os.Stderr, "No Raft storage directory specified\n")
		os.Exit(1)
	}
	peerFile := "./config.json"
	var peers []Peer
	if data, err := ioutil.ReadFile(peerFile); err == nil {
		json.Unmarshal(data, &peers)
	} else {
		panic(err)
	}
	// Ensure Raft storage exists.
	raftDir := flag.Arg(0)
	if raftDir == "" {
		fmt.Fprintf(os.Stderr, "No Raft storage directory specified\n")
		os.Exit(1)
	}
	os.MkdirAll(raftDir, 0700)

	pygokv, err := New(peers, peers[nodeID-1].ID, raftDir, inmem)
	if err != nil {
		log.Fatalf("failed to open store: %s", err.Error())
	}

	h := NewHttpd(fmt.Sprintf("%s:%d", peers[nodeID-1].Host, peers[nodeID-1].Port+6000), pygokv.Store)
	if err := h.Start(); err != nil {
		log.Fatalf("failed to start HTTP service: %s", err.Error())
	}

	log.Println("hraftd started successfully")

	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, os.Interrupt)
	<-terminate
	log.Println("hraftd exiting")
}
