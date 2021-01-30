package postoffice

import (
	"reflect"
	"sync"

	"golang.org/x/sync/semaphore"
)

type reloadChan struct {
	ch   chan struct{}
	lock *semaphore.Weighted

	exit chan struct{}
}

func newReloadChan() *reloadChan {
	return &reloadChan{
		ch:   make(chan struct{}),
		lock: semaphore.NewWeighted(1),
	}
}

func (ch *reloadChan) fire() {
	if ch.lock.TryAcquire(1) {
		select {
		case ch.ch <- struct{}{}:
		case <-ch.exit:
		}
	}
}

func (ch *reloadChan) reset() {
	if !ch.lock.TryAcquire(1) {
		ch.lock.Release(1)
	}
}

func (ch *reloadChan) toSelectCase() reflect.SelectCase {
	return reflect.SelectCase{
		Dir:  reflect.SelectRecv,
		Chan: reflect.ValueOf(ch.ch),
	}
}

type reloads struct {
	reloadChans sync.Map
}

func (r *reloads) Add(reCh *reloadChan) {
	r.reloadChans.Store(reCh, nil)
}

func (r *reloads) Remove(reCh *reloadChan) {
	r.reloadChans.Delete(reCh)
}

func (r *reloads) Reload() {
	r.reloadChans.Range(func(key interface{}, value interface{}) bool {
		reCh, _ := key.(*reloadChan)
		reCh.fire()
		return false
	})
}
