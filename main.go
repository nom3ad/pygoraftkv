package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"

	httpd "github.com/otoolep/hraftd/http"
	"github.com/otoolep/hraftd/store"
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
	var peers []store.Peer
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

	s := store.New(inmem)
	s.RaftDir = raftDir
	s.RaftBind = fmt.Sprintf("%s:%d", peers[nodeID-1].Host, peers[nodeID-1].Port)
	if err := s.Open(peers[nodeID-1].ID, peers); err != nil {
		log.Fatalf("failed to open store: %s", err.Error())
	}

	h := httpd.New(fmt.Sprintf("%s:%d", peers[nodeID-1].Host, peers[nodeID-1].Port+6000), s)
	if err := h.Start(); err != nil {
		log.Fatalf("failed to start HTTP service: %s", err.Error())
	}

	log.Println("hraftd started successfully")

	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, os.Interrupt)
	<-terminate
	log.Println("hraftd exiting")
}

// func join(joinAddr, raftAddr, nodeID string) error {
// 	b, err := json.Marshal(map[string]string{"addr": raftAddr, "id": nodeID})
// 	if err != nil {
// 		return err
// 	}
// 	resp, err := http.Post(fmt.Sprintf("http://%s/join", joinAddr), "application-type/json", bytes.NewReader(b))
// 	if err != nil {
// 		return err
// 	}
// 	defer resp.Body.Close()

// 	return nil
// }
