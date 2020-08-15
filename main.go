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
	var lock *sync.Mutex

	g.outer.Lock()
	shouldCollect := g.shouldCollect()
	if !shouldCollect {
		lock = g.acquire(id)
	}
	g.outer.Unlock()
	if shouldCollect {
		g.collect()
		g.outer.Lock()
		lock = g.acquire(id)
		g.outer.Unlock()
	}
	lock.Lock()
}

func (g *lockGroup) unlock(id string) {
	g.outer.Lock()
	lock := g.acquire(id)
	g.outer.Unlock()
	lock.Unlock()
}

func (g *lockGroup) shouldCollect() bool {
	return len(g.locks) == maxLocks
}

func (g *lockGroup) collect() {
	println("Collecting...")
	var locks []*sync.Mutex
	g.outer.Lock()
	if len(g.locks) == maxLocks {
		for _, v := range g.locks {
			locks = append(locks, v)
		}
	}
	g.outer.Unlock()
	for _, v := range locks {
		v.Lock()
	}
	g.outer.Lock()
	g.locks = make(map[string]*sync.Mutex)
	g.outer.Unlock()
	println("Collected")
}

func (g *lockGroup) acquire(id string) *sync.Mutex {
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

	for i := 0; i < 5*maxLocks; i++ {
		wg.Add(1)
		go handler(&wg, strconv.Itoa(i))
	}

	wg.Wait()
}
