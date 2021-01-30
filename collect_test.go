package postoffice

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/suite"
)

////////// Suite //////////

type CollectSuite struct {
	poSuite
}

func TestCollectSuite(t *testing.T) {
	suite.Run(t, new(CollectSuite))
}

////////// Testing //////////

func (s *CollectSuite) TestNoneTrue() {
	mail := s.po.Collect(context.Background(), false)

	// Verification
	s.Assert().Empty(mail)
}

func (s *CollectSuite) TestTrue() {
	// Setup
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		slot := s.po.getSlot("test")
		wg.Done()
		slot <- ""
	}()
	wg.Wait()

	mail := s.po.Collect(context.Background(), false)

	// Verification
	s.Require().NotEmpty(mail)
	s.Assert().EqualValues(&Mail{"test", ""}, mail[0])
}

func (s *CollectSuite) TestMultiTrue() {
	// Setup
	wg := sync.WaitGroup{}
	f := func(address string) {
		slot := s.po.getSlot(address)
		wg.Done()
		slot <- ""
	}

	wg.Add(2)
	go f("test")
	go f("nottest")

	wg.Wait()

	mail := s.po.Collect(context.Background(), false)

	// Verification
	s.Assert().Len(mail, 2)
}

func (s *CollectSuite) TestMultiPassNoneTrue() {
	mail := s.po.Collect(context.Background(), true)

	// Verification
	s.Assert().Empty(mail)
}

func (s *CollectSuite) TestMultiPassTrue() {
	// Setup
	wgStart := sync.WaitGroup{}
	wgCancel := sync.WaitGroup{}
	f := func(address string) {
		defer wgCancel.Done()
		slot := s.po.getSlot(address)
		wgStart.Done()
		slot <- ""
	}

	wgStart.Add(2)
	wgCancel.Add(2)
	go f("test")
	go f("test")

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		wgCancel.Wait()
		cancel()
	}()

	wgStart.Wait()
	mail := s.po.Collect(ctx, true)

	// Verification
	s.Require().Len(mail, 2)
	s.Assert().EqualValues(&Mail{"test", ""}, mail[0])
	s.Assert().EqualValues(&Mail{"test", ""}, mail[1])
}

func (s *CollectSuite) TestMultiSlotMultiPassTrue() {
	// Setup
	wgStart := sync.WaitGroup{}
	wgCancel := sync.WaitGroup{}
	f := func(address string) {
		defer wgCancel.Done()
		slot := s.po.getSlot(address)
		wgStart.Done()
		slot <- ""
	}

	wgStart.Add(4)
	wgCancel.Add(4)
	go f("test")
	go f("test")
	go f("nottest")
	go f("nottest")

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		wgCancel.Wait()
		cancel()
	}()

	wgStart.Wait()
	mail := s.po.Collect(ctx, true)

	// Verification
	s.Assert().Len(mail, 4)
}

func (s *CollectSuite) TestFalse() {
	// Setup
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	mail := s.po.Collect(ctx, true)

	// Verification
	s.Assert().Empty(mail)
}

func (s *CollectSuite) TestMultiPassFalse() {
	// Setup
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	mail := s.po.Collect(ctx, true)

	// Verification
	s.Assert().Empty(mail)
}
