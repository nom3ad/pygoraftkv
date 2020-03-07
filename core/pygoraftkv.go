package core

import (
	"fmt"
)

type PyGoRaftKV struct {
	*Store
	Quorum []Peer
}

func New(quorum []Peer, myId string, raftDir string, inmem bool) (*PyGoRaftKV, error) {

	s := NewStore(inmem)
	s.RaftDir = raftDir
	var me Peer
	for _, p := range quorum {
		if p.ID == myId {
			me = p
			break
		}
	}
	if me.ID != myId {
		return nil, fmt.Errorf("self ID: %s is not found in quorum", myId)
	}
	s.RaftBind = fmt.Sprintf("%s:%d", me.Host, me.Port)
	if err := s.Open(myId, quorum); err != nil {
		return nil, err
	}
	pyGoRaftKv := PyGoRaftKV{
		Store:  s,
		Quorum: quorum,
	}

	return &pyGoRaftKv, nil
}
