package fabric

// Basic imports
import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// UtilsTestSuite -
type UtilsTestSuite struct {
	suite.Suite
}

func (suite *UtilsTestSuite) SetupTest() {
}

func (suite *UtilsTestSuite) TestGenerateReqID() {
	id1 := generateReqID()
	id2 := generateReqID()
	suite.Assert().NotEqual(id1, id2)
}

func TestUtilsTestSuite(t *testing.T) {
	suite.Run(t, new(UtilsTestSuite))
}
