package pool

import (
	"sync"
	"sync/atomic"
)

// MetricsPool is a sync.Pool that records the number of gets and puts
type MetricsPool struct {
	P sync.Pool
	i int32
}

// Put adds x to the pool
func (p *MetricsPool) Put(x interface{}) {
	atomic.AddInt32(&p.i, -1)
	p.P.Put(x)
}

// Get selects an arbitrary item from the Pool, removes it from the
// Pool, and returns it to the caller.
// Get may choose to ignore the pool and treat it as empty.
// Callers should not assume any relation between values passed to Put and
// the values returned by Get.
//
// If Get would otherwise return nil and p.New is non-nil, Get returns
// the result of calling p.New.
func (p *MetricsPool) Get() interface{} {
	atomic.AddInt32(&p.i, 1)
	return p.P.Get()
}

// Releases returns the current # of items requested from the pool, but not returned.
//
// This is mainly for detecting leaks
func (p *MetricsPool) Releases() int32 {
	return atomic.LoadInt32(&p.i)
}
