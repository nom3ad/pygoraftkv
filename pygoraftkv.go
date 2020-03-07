package main

import (
	"fmt"

	"github.com/nom3ad/pygoraftkv/store"
)

type PyGoRaftKV struct {
	*store.Store
	Quorum []store.Peer
}

func New(quorum []store.Peer, myId string, raftDir string, inmem bool) (*PyGoRaftKV, error) {

	s := store.New(inmem)
	s.RaftDir = raftDir
	s.RaftBind = fmt.Sprintf("%s:%d", quorum[nodeID-1].Host, quorum[nodeID-1].Port)
	var me store.Peer
	for _, p := range quorum {
		if p.ID == myId {
			me = p
			break
		}
	}
	if me.ID != myId {
		return nil, fmt.Errorf("self ID: %s is not found in quorum", myId)
	}
	if err := s.Open(myId, quorum); err != nil {
		return nil, err
	}
	pyGoRaftKv := PyGoRaftKV{
		Store:  s,
		Quorum: quorum,
	}

	return &pyGoRaftKv, nil
}
