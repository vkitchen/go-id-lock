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

func (g *lockGroup) lock(id string) {
	g.outer.Lock()
	lock := g.acquire(id)
	g.outer.Unlock()
	lock.Lock()
}

func (g *lockGroup) unlock(id string) {
	g.outer.Lock()
	lock := g.acquire(id)
	g.outer.Unlock()
	lock.Unlock()
}

func (g *lockGroup) acquire(id string) *sync.Mutex {
	if len(g.locks) == maxLocks {
		for _, v := range g.locks {
			v.Lock()
		}
		g.locks = make(map[string]*sync.Mutex)
	}
	if lock, ok := g.locks[id]; ok {
		return lock
	}
	lock := &sync.Mutex{}
	g.locks[id] = lock
	return lock
}

var idLocks lockGroup

func init() {
	idLocks = newLockGroup()
}

func handler(wg *sync.WaitGroup, id string) {
	defer wg.Done()

	idLocks.lock(id)
	println("Locked", id)
	idLocks.unlock(id)
	println("Unlocked", id)
}

func main() {
	var wg sync.WaitGroup

	for i := 0; i < maxLocks; i++ {
		wg.Add(1)
		go handler(&wg, strconv.Itoa(i))
	}

	wg.Wait()
}
