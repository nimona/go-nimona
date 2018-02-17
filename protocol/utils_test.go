package protocol

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

func TestUtilsTestSuite(t *testing.T) {
	suite.Run(t, new(UtilsTestSuite))
}
