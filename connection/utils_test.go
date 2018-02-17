package connection

// Basic imports
import (
	"io"
	"sync"
	"testing"

	suite "github.com/stretchr/testify/suite"
)

// UtilsTestSuite -
type UtilsTestSuite struct {
	suite.Suite
}

func (suite *UtilsTestSuite) SetupTest() {
}

func (suite *UtilsTestSuite) TestReadWriteToken() {
	reader, writter := io.Pipe()
	payload := []byte("hello")

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		retPayload, err := ReadToken(reader)
		suite.Assert().Nil(err)
		suite.Assert().Equal(payload, retPayload)
		wg.Done()
	}()

	err := WriteToken(writter, payload)
	suite.Assert().Nil(err)
	wg.Wait()
}

func TestUtilsTestSuite(t *testing.T) {
	suite.Run(t, new(UtilsTestSuite))
}
