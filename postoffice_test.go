package postoffice

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

////////// Base Suite //////////

type poSuite struct {
	suite.Suite
	po *PostOffice
}

func (s *poSuite) SetupSuite() {}

func (s *poSuite) SetupTest() {
	s.po = &PostOffice{}
}

func (s *poSuite) TearDownTest() {}

func (s *poSuite) TearDownSuite() {}

////////// Suite //////////

type HelperSuite struct {
	poSuite
}

func TestHelperSuite(t *testing.T) {
	suite.Run(t, new(HelperSuite))
}

////////// Testing //////////

///// Addresses /////
func (s *HelperSuite) TestAddresses() {
	// Setup
	s.po.getSlot("a")
	s.po.getSlot("b")

	// Verification
	addrs := s.po.Addresses()
	s.Require().NotEmpty(addrs)
	s.Assert().Len(addrs, 2)
}

///// selectCases /////

///// getSlot /////

func (s *HelperSuite) TestgetSlot() {
	// Setup
	slot := s.po.getSlot("test")

	// Verification
	s.Require().NotNil(slot)
	s.Assert().Empty(slot)
}
