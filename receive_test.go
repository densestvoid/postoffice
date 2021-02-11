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

type ReceiveSuite struct {
	poSuite
}

func TestReceiveSuite(t *testing.T) {
	suite.Run(t, new(ReceiveSuite))
}

////////// Testing //////////

func (s *ReceiveSuite) TestTrue() {
	// Setup
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		slot := s.po.getSlot("test")
		slot <- ""
	}()

	mail, received := s.po.Receive(context.Background(), "test")

	wg.Wait()

	// Verification
	s.Require().True(received)
	s.Require().NotNil(mail)
	s.Assert().EqualValues(&Mail{"test", ""}, mail)
}

func (s *ReceiveSuite) TestMultiTrue() {
	// Setup
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		slot := s.po.getSlot("test")
		slot <- ""
	}()

	mail, received := s.po.Receive(context.Background(), []interface{}{"test", "nottest"}...)

	wg.Wait()

	// Verification
	s.Require().True(received)
	s.Require().NotNil(mail)
	s.Assert().EqualValues(&Mail{"test", ""}, mail)
}

func (s *ReceiveSuite) TestBroadcastTrue() {
	// Setup
	wg := &sync.WaitGroup{}
	defer wg.Wait()

	f := func(name string) {
		defer wg.Done()

		slot := s.po.getSlot(name)

		timer := time.NewTimer(time.Second)
		select {
		case slot <- name:
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
			mail, received := s.po.Receive(ctx)
			if !received {
				break
			}

			count++

			s.Require().NotNil(mail)
			s.Require().NotNil(mail.Contents)
			_, isString := mail.Contents.(string)
			s.Assert().True(isString)
		}
	}()
	s.Assert().Eventually(func() bool { return count == iterations }, time.Second*5, time.Second/10)
}

func (s *ReceiveSuite) TestFalse() {
	// Setup
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, received := s.po.Receive(ctx, "test")

	// Verification
	s.Require().False(received)
}

func (s *ReceiveSuite) TestMultiFalse() {
	// Setup
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, received := s.po.Receive(ctx, "test", "nottest")

	// Verification
	s.Require().False(received)
}

func (s *ReceiveSuite) TestBroadcastFalse() {
	// Setup
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, received := s.po.Receive(ctx)

	// Verification
	s.Require().False(received)
}
