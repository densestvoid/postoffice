package postoffice

import (
	"reflect"
	"sync"
)

type PostOffice struct {
	slots   sync.Map
	reloads reloads

	closeOnce sync.Once
	close     chan struct{}
}

func (po *PostOffice) Close() {
	po.closeOnce.Do(func() {
		po.close = make(chan struct{})
		close(po.close)
	})
}

func (po *PostOffice) selectCases() ([]string, []reflect.SelectCase) {
	var keys []string
	var cases []reflect.SelectCase
	po.slots.Range(func(key interface{}, value interface{}) bool {
		keyStr, ok := key.(string)
		if !ok {
			return true
		}

		keys = append(keys, keyStr)

		ch := po.getSlot(keyStr)
		cases = append(cases, reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(ch),
		})
		return true
	})
	return keys, cases
}

func (po *PostOffice) Receive() (interface{}, interface{}, bool) {
	reCh := &reloadChan{ch: make(chan struct{})}
	po.reloads.Add(reCh)
	defer po.reloads.Remove(reCh)

	keys, cases := po.selectCases()
	cases = append(cases, reCh.SelectCase(), reflect.SelectCase{
		Dir:  reflect.SelectRecv,
		Chan: reflect.ValueOf(po.close),
	})

	index, value, _ := reflect.Select(cases)
	for index >= len(keys) {
		if index > len(keys) {
			return "", nil, false
		}

		reCh.Reset()

		keys, cases = po.selectCases()
		cases = append(cases, reCh.SelectCase(), reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(po.close),
		})

		index, value, _ = reflect.Select(cases)
	}

	return keys[index], value.Interface(), true
}

func (po *PostOffice) ReceiveFrom(slot interface{}) (interface{}, bool) {
	select {
	case <-po.close:
		return nil, false
	case value := <-po.getSlot(slot):
		return value, true
	}
}

func (po *PostOffice) Send(slot interface{}, i interface{}) bool {
	select {
	case <-po.close:
		return false
	case po.getSlot(slot) <- i:
		return true
	}
}

func (po *PostOffice) getSlot(slot interface{}) chan interface{} {
	value, loaded := po.slots.LoadOrStore(slot, make(chan interface{}))
	ch, _ := value.(chan interface{})
	if !loaded {
		po.reloads.Reload()
	}
	return ch
}
