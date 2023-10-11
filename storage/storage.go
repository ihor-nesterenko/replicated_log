package storage

import "sync"

var inMemoryList *List

func InitList() {
	inMemoryList = &List{
		list:    []string{},
		RWMutex: sync.RWMutex{},
	}
}

func GetList() *List {
	return inMemoryList
}

// List is concurrent safe List
type List struct {
	list []string
	sync.RWMutex
}

func NewList() List {
	return List{
		list:    []string{},
		RWMutex: sync.RWMutex{},
	}
}

// Get gets stored List
func (l *List) Get() []string {
	l.RLock()
	defer l.RUnlock()

	return l.list
}

// Add adds new element to the List
func (l *List) Add(elem string) {
	l.Lock()
	defer l.Unlock()

	l.list = append(l.list, elem)
}
