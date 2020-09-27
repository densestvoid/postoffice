package postoffice

import (
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/suite"
)

////////// Suite //////////

type PostOfficeSuite struct {
	suite.Suite
	po *PostOffice
}

func (s *PostOfficeSuite) SetupSuite() {}

func (s *PostOfficeSuite) SetupTest() {
	s.po = &PostOffice{}
}

func (s *PostOfficeSuite) TearDownTest() {
	s.po.Close()
}

func (s *PostOfficeSuite) TearDownSuite() {}

func TestPostOfficeSuite(t *testing.T) {
	suite.Run(t, new(PostOfficeSuite))
}

////////// Testing //////////

///// Receive /////

func (s *PostOfficeSuite) TestReceive_NotNil() {
	// Setup
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		slot := s.po.getSlot("test")
		slot <- ""
	}()

	key, value, received := s.po.Receive()

	wg.Wait()

	// Verification
	s.Require().True(received)
	s.Require().NotNil(value)
	_, isString := value.(string)
	s.Assert().True(isString)
	s.Assert().EqualValues("test", key)
}

func (s *PostOfficeSuite) TestReceive_Nil() {
	// Setup
	s.po.Close()

	// Verification
	_, _, received := s.po.Receive()
	s.Require().False(received)
}

func (s *PostOfficeSuite) TestReceive_Multi() {
	// Setup
	wg := sync.WaitGroup{}

	f := func(name string) {
		defer wg.Done()
		slot := s.po.getSlot(name)
		slot <- name
	}

	var iterations = 5

	wg.Add(iterations)
	for i := 0; i < iterations; i++ {
		go f(strconv.Itoa(i))
	}

	// Verification
	var count int
	for _, value, received := s.po.Receive(); received; _, value, received = s.po.Receive() {
		count++
		s.Require().NotNil(value)
		_, isString := value.(string)
		s.Assert().True(isString)
	}
	s.Assert().EqualValues(iterations, count)

	wg.Wait()
}

///// ReceiveFrom /////

func (s *PostOfficeSuite) TestReceiveFrom_NotNil() {
	// Setup
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		slot := s.po.getSlot("test")
		slot <- ""
	}()

	value, received := s.po.ReceiveFrom("test")

	wg.Wait()

	// Verification
	s.Require().True(received)
	s.Require().NotNil(value)
	_, isString := value.(string)
	s.Assert().True(isString)
}

func (s *PostOfficeSuite) TestReceiveFrom_Nil() {
	// Setup
	s.po.Close()

	// Verification
	_, received := s.po.ReceiveFrom("test")
	s.Require().False(received)
}

///// Send /////

func (s *PostOfficeSuite) TestSend_True() {
	// Setup
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.Require().True(s.po.Send("test", ""))
	}()

	slot := s.po.getSlot("test")
	msg := <-slot

	wg.Wait()

	// Verification
	s.Require().NotNil(msg)
	_, isString := msg.(string)
	s.Assert().True(isString)
}

func (s *PostOfficeSuite) TestSend_False() {
	// Setup
	s.po.Close()

	// Verification
	s.Require().False(s.po.Send("test", ""))
}

func (s *PostOfficeSuite) TestgetSlot() {
	// Setup
	slot := s.po.getSlot("test")

	// Verification
	s.Require().NotNil(slot)
	s.Assert().Empty(slot)
}
