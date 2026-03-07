package proxy

import (
	"sync"
	"sync/atomic"
)

type LoadBalancer struct {
	counters sync.Map
}

func NewLoadBalancer() *LoadBalancer {
	return &LoadBalancer{}
}

func (lb *LoadBalancer) NextIndex(key string, size int) int {
	if size <= 1 {
		return 0
	}

	counter, _ := lb.counters.LoadOrStore(key, &atomic.Uint64{})
	idx := counter.(*atomic.Uint64).Add(1) - 1
	return int(idx % uint64(size))
}
