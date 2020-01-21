package counter

import "sync"

type Counter struct {
	Counter int
	*sync.RWMutex
}

func CreateCounter() *Counter {
	return &Counter{
		Counter: 0,
		RWMutex: new(sync.RWMutex),
	}
}

func (counter *Counter) increase() {
	counter.Lock()
	defer counter.Unlock()

	counter.Counter++
}

func (counter *Counter) Current() int {
	counter.RLock()
	defer counter.RUnlock()

	return counter.Counter
}

func (counter *Counter) Next() int {
	counter.increase()
	return counter.Current()
}
