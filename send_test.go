package postoffice

import (
	"context"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

////////// Suite //////////

type SendSuite struct {
	poSuite
}

func TestSendSuite(t *testing.T) {
	suite.Run(t, new(SendSuite))
}

////////// Testing //////////

func (s *SendSuite) TestTrue() {
	// Setup
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, sent := s.po.Send(context.Background(), "", "test")
		s.Require().True(sent)
	}()

	slot := s.po.getSlot("test")
	msg := <-slot

	wg.Wait()

	// Verification
	s.Require().NotNil(msg)
	_, isString := msg.(string)
	s.Assert().True(isString)
}

func (s *SendSuite) TestMutliTrue() {
	// Setup
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, sent := s.po.Send(context.Background(), "", "test", "nottest")
		s.Require().True(sent)
	}()

	slot := s.po.getSlot("test")
	msg := <-slot

	wg.Wait()

	// Verification
	s.Require().NotNil(msg)
	_, isString := msg.(string)
	s.Assert().True(isString)
}

func (s *SendSuite) TestBroadcastTrue() {
	// Setup
	wg := &sync.WaitGroup{}
	defer wg.Wait()

	f := func(name string) {
		defer wg.Done()

		slot := s.po.getSlot(name)

		timer := time.NewTimer(time.Second)
		select {
		case <-slot:
		case <-timer.C:
		}
	}

	var iterations = 5

	wg.Add(iterations)
	for i := 0; i < iterations; i++ {
		go f(strconv.Itoa(i))
	}

	// Verification
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var count int
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			_, sent := s.po.Send(ctx, "")
			if !sent {
				break
			}

			count++
		}
	}()
	s.Assert().Eventually(func() bool { return count == iterations }, time.Second*5, time.Second/10)
}

func (s *SendSuite) TestFalse() {
	// Setup
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Verification
	_, sent := s.po.Send(ctx, "", "test")
	s.Require().False(sent)
}

func (s *SendSuite) TestMultiFalse() {
	// Setup
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, sent := s.po.Send(ctx, "", "test", "nottest")

	// Verification
	s.Require().False(sent)
}

func (s *SendSuite) TestBroadcastFalse() {
	// Setup
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, sent := s.po.Send(ctx, "")

	// Verification
	s.Require().False(sent)
}
