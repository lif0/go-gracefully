package main

import "sync"

type memStore struct {
	data []string
	mu   sync.Mutex
}

func newSimpleMemStore() *memStore {
	return &memStore{}
}

func (s *memStore) Store(data []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = append(s.data, data...)
}

func (s *memStore) GetAndClear() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	data := s.data
	s.data = nil
	return data
}
