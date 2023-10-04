package storage

import "sync"

// FIXME: singletone is bad. I'll will refactor it to something normal in the future, I promise
var inMemoryList *list

func InitList() {
	inMemoryList = &list{
		list:    []string{},
		RWMutex: sync.RWMutex{},
	}
}

func GetList() *list {
	return inMemoryList
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
