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

///// selectCases /////

///// getSlot /////

func (s *HelperSuite) TestgetSlot() {
	// Setup
	slot := s.po.getSlot("test")

	// Verification
	s.Require().NotNil(slot)
	s.Assert().Empty(slot)
}
