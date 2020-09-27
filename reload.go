package postoffice

import (
	"reflect"
	"sync"
)

type reloadChan struct {
	ch    chan struct{}
	close sync.Mutex
}

func (ch *reloadChan) Fire() {
	ch.close.Lock()
	close(ch.ch)
}

func (ch *reloadChan) Reset() {
	ch.ch = make(chan struct{})
	ch.close.Unlock()
}

func (ch *reloadChan) SelectCase() reflect.SelectCase {
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
		reCh.Fire()
		return false
	})
}
