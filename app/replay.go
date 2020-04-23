package app

import (
	"sync"
	"time"
)

type bucket map[string]bool
type Replay struct {
	interval time.Duration
	size     uint
	mtx      sync.RWMutex
	buckets  []bucket
}

func NewReplay(interval time.Duration, size uint) *Replay {
	r := new(Replay)
	r.interval = interval
	r.size = size
	r.buckets = []bucket{make(bucket)}
	go r.start()
	return r
}

func (r *Replay) start() {
	ticker := time.NewTicker(r.interval)
	for range ticker.C {
		r.mtx.Lock()
		r.buckets = append([]bucket{make(bucket)}, r.buckets...)
		if len(r.buckets) > int(r.size) {
			r.buckets = r.buckets[:r.size]
		}
		r.mtx.Unlock()
	}
}

func (r *Replay) Dupe(hash string) bool {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	for _, b := range r.buckets {
		if b[hash] {
			return true
		}
	}
	r.buckets[0][hash] = true
	return false
}
