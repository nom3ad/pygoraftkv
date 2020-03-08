package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/signal"
	"reflect"
	"strconv"
	"time"

	"github.com/nom3ad/pygoraftkv/pygoraftkv"
	"github.com/nom3ad/pygoraftkv/rpc"
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
	peerFile := "./config-single.json"
	var quorum []pygoraftkv.Member
	if data, err := ioutil.ReadFile(peerFile); err == nil {
		json.Unmarshal(data, &quorum)
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
	conn, err := net.Dial("tcp", "127.0.0.1:50000")
	if err != nil {
		log.Fatalf("failed to dial rpc client: %s", err.Error())
		return
	}

	client := rpc.NewSession(conn, true)
	getter := func(key string) (string, error) {
		retval, xerr := client.Send("Get", key)
		if xerr != nil {
			return "", xerr
		}
		if retval.Kind() != reflect.String {
			return "", fmt.Errorf("Invalid retiurn type")
		}
		return retval.String(), nil
	}
	setter := func(key string, val string) error {
		_, xerr := client.Send("Set", key, val)
		if xerr != nil {
			return xerr
		}

		return nil
	}
	deleter := func(key string) error {
		_, xerr := client.Send("Delete", key)
		if xerr != nil {
			return xerr
		}
		return nil
	}
	pygokv, err := pygoraftkv.NewPyGoKV(quorum, quorum[nodeID-1].ID, raftDir, inmem, getter, setter, deleter)
	if err != nil {
		log.Fatalf("failed to open store: %s", err.Error())
	}

	pygokv.Open()

	// h := pygoraftkv.NewHttpd(fmt.Sprintf("%s:%d", peers[nodeID-1].Host, peers[nodeID-1].Port+6000), pygokv.Store)
	// if err := h.Start(); err != nil {
	// 	log.Fatalf("failed to start HTTP service: %s", err.Error())
	// }

	log.Println("hraftd started successfully")

	ticker := time.NewTicker(time.Second * 5)

	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, os.Interrupt)

	var k int
	for {
		select {
		case <-terminate:
			log.Println("hraftd exiting")
			return
		case dt := <-ticker.C:
			k++
			if err := pygokv.Set(strconv.Itoa(k), dt.String()); err != nil {
				panic(err)
			}
			if val, err := pygokv.Get(strconv.Itoa(k - 1)); err != nil {
				fmt.Println("GET-ERROR:", k-1, err)
			} else {
				fmt.Printf("Value of %d = %s\n", k-1, val)
			}
		}
	}

}
