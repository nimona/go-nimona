package fabric

// Basic imports
import (
	"bufio"
	"encoding/binary"
	"errors"
	"io"
	"sync"
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

func (suite *UtilsTestSuite) TestWriteTokenError() {
	reader, writter := io.Pipe()
	payload := []byte("hello")

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		_, err := ReadToken(reader)
		suite.Assert().NotNil(err)
		wg.Done()
	}()

	bw := bufio.NewWriter(writter)
	vb := make([]byte, 16)
	n := binary.PutUvarint(vb, uint64(len(payload)))
	wb := append(vb[:n], payload[:2]...)
	writter.Write(wb)
	bw.Flush()
	writter.CloseWithError(errors.New("error"))
	wg.Wait()
}
