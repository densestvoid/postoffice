package postoffice

import (
	"reflect"
	"sync"
)

type PostOffice struct {
	slots sync.Map

	closeOnce sync.Once
	close     chan struct{}
}

func (po *PostOffice) Close() {
	po.closeOnce.Do(func() {
		po.close = make(chan struct{})
		close(po.close)
	})
}

func (po *PostOffice) selectCases() []reflect.SelectCase {
	var cases []reflect.SelectCase
	po.slots.Range(func(key interface{}, value interface{}) bool {
		keyStr, ok := key.(string)
		if !ok {
			return true
		}

		ch := po.getSlot(keyStr)
		cases = append(cases, reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(ch),
		})
		return true
	})
	return cases
}

// TODO: Read the next mail
func (po *PostOffice) Receive() (string, interface{}, bool) {

}

// TODO: Read the next mail addresses to x
func (po *PostOffice) ReceiveFrom(slot string) (interface{}, bool) {
	select {
	case <-po.close:
		return nil, false
	case value := <-po.getSlot(slot):
		return value, true
	}
}

// TODO: Send mail addresses to x
func (po *PostOffice) Send(slot string, i interface{}) bool {
	select {
	case <-po.close:
		return false
	case po.getSlot(slot) <- i:
		return true
	}
}

func (po *PostOffice) getSlot(slot string) chan interface{} {
	value, _ := po.slots.LoadOrStore(slot, make(chan interface{}))
	ch, _ := value.(chan interface{})
	return ch
}
