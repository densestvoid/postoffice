package postoffice

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/suite"
)

////////// Suite //////////

type BroadcastSuite struct {
	poSuite
}

func TestBroadcastSuite(t *testing.T) {
	suite.Run(t, new(BroadcastSuite))
}

////////// Testing //////////

func (s *BroadcastSuite) TestNoneTrue() {
	mail := s.po.Broadcast(context.Background(), false)

	// Verification
	s.Assert().Empty(mail)
}

func (s *BroadcastSuite) TestTrue() {
	// Setup
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		slot := s.po.getSlot("test")
		wg.Done()
		<-slot
	}()
	wg.Wait()

	addresses := s.po.Broadcast(context.Background(), "")

	// Verification
	s.Require().NotEmpty(addresses)
	s.Assert().EqualValues("test", addresses[0])
}

func (s *BroadcastSuite) TestMultiTrue() {
	// Setup
	wg := sync.WaitGroup{}
	f := func(address string) {
		slot := s.po.getSlot(address)
		wg.Done()
		<-slot
	}

	wg.Add(2)
	go f("test")
	go f("nottest")

	wg.Wait()

	addresses := s.po.Broadcast(context.Background(), "")

	// Verification
	s.Assert().Len(addresses, 2)
}

func (s *BroadcastSuite) TestFalse() {
	// Setup
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	addresses := s.po.Broadcast(ctx, true)

	// Verification
	s.Assert().Empty(addresses)
}
