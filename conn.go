package nimona

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"sync/atomic"

	"nimona.io/internal/xsync"
)

type Conn struct {
	writerQueue *xsync.Queue[*pendingWrite]
	readerQueue *xsync.Queue[*pendingRead]

	requests   *xsync.Map[uint64, *pendingRequest]
	requestSeq uint64
}

type Message struct {
	Body  []byte
	Conn  *Conn
	Reply func([]byte) error
}

type pendingRequest struct {
	dst  []byte
	err  error
	done chan struct{}
}

type pendingWrite struct {
	seq  uint64
	buf  []byte
	wait bool
	err  error
	done chan struct{}
}

type pendingRead struct {
	seq  uint64
	body []byte
}

func NewConn() *Conn {
	c := &Conn{
		writerQueue: xsync.NewQueue[*pendingWrite](10),
		readerQueue: xsync.NewQueue[*pendingRead](10),
		requests:    xsync.NewMap[uint64, *pendingRequest](),
	}
	return c
}

func (c *Conn) Handle(conn net.Conn) error {
	writerDone := make(chan error)
	go func() {
		writerDone <- c.writeLoop(conn)
		close(writerDone)
	}()

	readerDone := make(chan error)
	go func() {
		readerDone <- c.readLoop(conn)
		close(readerDone)
	}()

	// TODO: needs testing and figuring out edge cases
	var err error
	select {
	case err = <-writerDone:
		close(readerDone)
	case err = <-readerDone:
		close(writerDone)
	}

	c.writerQueue.Close()

	return err
}

func (c *Conn) Request(payload []byte) ([]byte, error) {
	pr := &pendingRequest{
		done: make(chan struct{}),
	}

	requestSeq := c.next()
	c.requests.Store(requestSeq, pr)

	err := c.write(requestSeq, payload, false)
	if err != nil {
		close(pr.done)
		c.requests.Delete(requestSeq)
		return nil, err
	}

	<-pr.done
	return pr.dst, pr.err
}

func (c *Conn) next() uint64 {
	return atomic.AddUint64(&c.requestSeq, 1)
}

func (c *Conn) write(seq uint64, payload []byte, wait bool) error {
	pw := &pendingWrite{
		seq:  seq,
		buf:  payload,
		wait: wait,
		done: make(chan struct{}),
	}

	err := c.writerQueue.Push(pw)
	if err != nil {
		return fmt.Errorf("pushing to writer queue: %w", err)
	}

	if !wait {
		return nil
	}

	<-pw.done
	return pw.err
}

func (c *Conn) writeLoop(conn net.Conn) error {
	for {
		pw, err := c.writerQueue.Pop()
		if err != nil {
			if errors.Is(err, xsync.ErrQueueClosed) {
				return nil
			}
			return fmt.Errorf("pop from writer queue: %w", err)
		}

		// TODO: probably too expensive
		header := []byte{}
		header = binary.AppendUvarint(header, pw.seq)
		header = binary.AppendUvarint(header, uint64(len(pw.buf)))
		_, err = conn.Write(append(header, pw.buf...))
		if pw.wait {
			pw.err = err
			close(pw.done)
		}
		if err != nil {
			return fmt.Errorf("write error: %w", err)
		}
	}
}

func (c *Conn) readLoop(conn net.Conn) error {
	reader := bufio.NewReader(conn)
	for {
		seq, err := binary.ReadUvarint(reader)
		if err != nil {
			return fmt.Errorf("read seq error: %w", err)
		}

		size, err := binary.ReadUvarint(reader)
		if err != nil {
			return fmt.Errorf("read size error: %w", err)
		}

		data := make([]byte, size)
		n, err := io.ReadFull(reader, data)
		if err != nil {
			return fmt.Errorf("read data error: %w", err)
		}

		if n != len(data) {
			return fmt.Errorf("read data error: %w", io.ErrUnexpectedEOF)
		}

		pr, exists := c.requests.Load(seq)
		if exists {
			c.requests.Delete(seq)
		}

		if seq == 0 || !exists {
			err := c.readerQueue.Push(&pendingRead{
				seq:  seq,
				body: data,
			})
			if err != nil {
				if errors.Is(err, xsync.ErrQueueClosed) {
					return nil
				}
				return fmt.Errorf("handler encountered an error: %w", err)
			}
			continue
		}

		pr.dst = data
		close(pr.done)
	}
}

func (c *Conn) Read() (*Message, error) {
	pr, err := c.readerQueue.Pop()
	if err != nil {
		if errors.Is(err, xsync.ErrQueueClosed) {
			return nil, io.EOF
		}
		return nil, fmt.Errorf("pop from reader queue: %w", err)
	}
	return &Message{
		Body: pr.body,
		Conn: c,
		Reply: func(payload []byte) error {
			if pr.seq == 0 {
				return fmt.Errorf("message is not a request")
			}
			return c.write(pr.seq, payload, true)
		},
	}, nil
}
