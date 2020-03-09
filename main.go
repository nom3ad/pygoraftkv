package main

import (
	"encoding/json"
	"fmt"
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

func main() {
	var quorum []pygoraftkv.Member
	raftDir := os.Getenv("PGRKV_RAFTDIR")
	addr := os.Getenv("PGRKV_BRIDGE_ADDR")
	quorumData := os.Getenv("PGRKV_QUORUM")
	myId := os.Getenv("PGRKV_MYID")
	if raftDir == "" || addr == "" || quorumData == "" || myId == "" {
		log.Fatalln("Envoronement not set")
	}
	if err := json.Unmarshal([]byte(quorumData), &quorum); err != nil {
		log.Fatalln("Unmarshal failed")
	}
	// Ensure Raft storage exists.
	if raftDir == "" {
		log.Fatalln("No Raft storage directory specified")
	}
	if err := os.MkdirAll(raftDir, 0700); err != nil {
		log.Fatalln("Mkdir failed : Raft storage directory")
	}
	// conn, err := net.Dial("tcp", "127.0.0.1:50000")
	conn, err := net.Dial("unix", addr)
	if err != nil {
		log.Fatalf("failed to dial rpc client: %s\n", err.Error())
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
	pygokv, err := pygoraftkv.NewPyGoKV(quorum, myId, raftDir, inmem, getter, setter, deleter)
	if err != nil {
		log.Fatalf("failed to creare store: %s", err.Error())
	}
	fut, err := pygokv.Open()
	if err != nil || fut.Error() != nil {
		log.Fatalf("failed to open store: %s", err.Error())
	}

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
				fmt.Printf("GET-ERROR: %d: err=%v\n", k-1, err.Error())
			} else {
				fmt.Printf("Value of %d = %s\n", k-1, val)
			}
		}
	}

}
