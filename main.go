package main

import (
	"strconv"
	"sync"
)

const (
	maxLocks = 3
)

type lockGroup struct {
	outer *sync.Mutex
	locks map[string]*sync.Mutex
}

func newLockGroup() lockGroup {
	var group lockGroup
	group.outer = &sync.Mutex{}
	group.locks = make(map[string]*sync.Mutex)
	return group
}

var idLocks lockGroup

func init() {
	idLocks = newLockGroup()
}

func handler(wg *sync.WaitGroup, id string) {
	defer wg.Done()

	idLocks.outer.Lock()

	if len(idLocks.locks) > maxLocks {
		println("Collecting...")
		for _, lock := range idLocks.locks {
			lock.Lock()
		}
		idLocks.locks = make(map[string]*sync.Mutex)
		println("Collected")
	}

	lock, ok := idLocks.locks[id]
	if !ok {
		lock = &sync.Mutex{}
		idLocks.locks[id] = lock
	}

	lock.Lock()
	println("Locked", id)

	idLocks.outer.Unlock()

	lock.Unlock()
	println("Unlocked", id)
}

func main() {
	var wg sync.WaitGroup

	for j := 0; j < 100; j++ {
		for i := 0; i < 5*maxLocks; i++ {
			wg.Add(1)
			go handler(&wg, strconv.Itoa(i))
		}
	}

	wg.Wait()
}
