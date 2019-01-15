package store

import (
	"encoding/json"

	"github.com/krantius/logging"
	"github.com/krantius/raft"
)

type Store interface {
	raft.Store
	Dump() []byte
}

type InMemory struct {
	m map[string]string
}

func New() *InMemory {
	return &InMemory{
		m: make(map[string]string),
	}
}

func (s *InMemory) Set(key string, val []byte) {
	logging.Tracef("Set called for key %s", key)
	s.m[key] = string(val)
}

func (s *InMemory) Delete(key string) {
	logging.Tracef("Delete called for key %s", key)
	delete(s.m, key)
}

func (s *InMemory) Dump() []byte {
	b, err := json.Marshal(s.m)
	if err != nil {
		panic(err)
	}

	return b
}
