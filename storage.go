package main

import "sync"

var inMemoryList list

func initList() {
	inMemoryList = list{
		list:    []string{},
		RWMutex: sync.RWMutex{},
	}
}

// list is concurrent safe list
type list struct {
	list []string
	sync.RWMutex
}

// Get gets stored list
func (l *list) Get() []string {
	l.RLock()
	defer l.RUnlock()

	return l.list
}

// Add adds new element to the list
func (l *list) Add(elem string) {
	l.Lock()
	defer l.Unlock()

	l.list = append(l.list, elem)
}
