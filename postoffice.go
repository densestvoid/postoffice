// Package postoffice is a combination channel fan-in/fan-out multiplexer
package postoffice

import (
	"context"
	"reflect"
	"sync"
)

type Mail struct {
	Address, Contents interface{}
}

type PostOffice struct {
	slots   sync.Map
	reloads reloads
}

func (po *PostOffice) selectCases(dir reflect.SelectDir, addresses ...interface{}) ([]interface{}, []reflect.SelectCase) {
	var keys []interface{}
	var cases []reflect.SelectCase

	if len(addresses) > 0 {
		for _, address := range addresses {
			keys = append(keys, address)

			ch := po.getSlot(address)
			cases = append(cases, reflect.SelectCase{
				Dir:  dir,
				Chan: reflect.ValueOf(ch),
			})
		}
	} else {
		po.slots.Range(func(key interface{}, value interface{}) bool {
			keys = append(keys, key)

			ch := po.getSlot(key)
			cases = append(cases, reflect.SelectCase{
				Dir:  dir,
				Chan: reflect.ValueOf(ch),
			})
			return true
		})
	}

	return keys, cases
}

func (po *PostOffice) Receive(ctx context.Context, addresses ...interface{}) (*Mail, bool) {
	reCh := newReloadChan()
	po.reloads.Add(reCh)
	defer po.reloads.Remove(reCh)

	for {
		keys, cases := po.selectCases(reflect.SelectRecv, addresses...)
		cases = append(cases, reCh.toSelectCase(), reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(ctx.Done()),
		})

		index, value, ok := reflect.Select(cases)

		if !ok && index == len(cases)-1 {
			return nil, false
		} else if index == len(cases)-2 {
			reCh.reset()
			continue
		}

		return &Mail{keys[index], value.Interface()}, true
	}
}

func (po *PostOffice) Collect(ctx context.Context, multiPass bool) []*Mail {
	var (
		wg                sync.WaitGroup
		collectedMailLock sync.Mutex
		collectedMail     []*Mail
	)

	po.slots.Range(func(key, contents interface{}) bool {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				mail, ok := po.Receive(ctx, key)
				if !ok {
					break
				}

				collectedMailLock.Lock()
				collectedMail = append(collectedMail, mail)
				collectedMailLock.Unlock()

				if !multiPass {
					break
				}
			}
		}()
		return true
	})

	wg.Wait()
	return collectedMail
}

func (po *PostOffice) Send(ctx context.Context, contents interface{}, addresses ...interface{}) (interface{}, bool) {
	reCh := newReloadChan()
	po.reloads.Add(reCh)
	defer po.reloads.Remove(reCh)

	for {
		keys, cases := po.selectCases(reflect.SelectSend, addresses...)
		for i := range cases {
			cases[i].Send = reflect.ValueOf(contents)
		}
		cases = append(cases, reCh.toSelectCase(), reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(ctx.Done()),
		})

		index, _, ok := reflect.Select(cases)

		if !ok && index == len(cases)-1 {
			return nil, false
		} else if index == len(cases)-2 {
			reCh.reset()
			continue
		}

		return keys[index], true
	}
}

func (po *PostOffice) Broadcast(ctx context.Context, contents interface{}) []interface{} {
	var (
		wg            sync.WaitGroup
		addressesLock sync.Mutex
		addresses     []interface{}
	)

	po.slots.Range(func(key, contents interface{}) bool {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, ok := po.Send(ctx, contents, key); ok {
				addressesLock.Lock()
				defer addressesLock.Unlock()

				addresses = append(addresses, key)
			}
		}()
		return true
	})

	wg.Wait()
	return addresses
}

func (po *PostOffice) getSlot(address interface{}) chan interface{} {
	value, loaded := po.slots.LoadOrStore(address, make(chan interface{}))
	ch, _ := value.(chan interface{})
	if !loaded {
		go po.reloads.Reload()
	}
	return ch
}
