package util

import (
	"sync"
)

// StringSet is a threadsafe set of strings
type StringSet struct {
	set  map[string]struct{}
	lock sync.RWMutex
}

// NewStringSet returns a StringSet popluated by the contents of "set"
func NewStringSet(set []string) *StringSet {
	s := &StringSet{
		lock: sync.RWMutex{},
	}
	s.Set(set)
	return s
}

// Contains returns true if the given string is contained in the StringSet
func (s *StringSet) Contains(v string) bool {
	s.lock.RLock()
	_, ok := s.set[v]
	s.lock.RUnlock()
	return ok
}

// ForEach calls the given function for each value in the StringSet
func (s *StringSet) ForEach(f func(v string)) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	for v := range s.set {
		f(v)
	}
}

// Set replaces the values in the StringSet with the new slice of strings
func (s *StringSet) Set(set []string) {
	s.lock.Lock()
	setMap := make(map[string]struct{})
	for _, v := range set {
		setMap[v] = struct{}{}
	}
	s.set = setMap
	s.lock.Unlock()
}
