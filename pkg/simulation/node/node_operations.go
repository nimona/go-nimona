package node

import (
	"bufio"
	"io"

	"nimona.io/pkg/context"
	"nimona.io/pkg/errors"
)

// Logs filters the node json logs for the specific tag
func (n *Node) Logs() (chan string, chan error) {
	logCh := make(chan string)
	errCh := make(chan error)

	go func() {
		ctx := context.Background()

		rdr, err := n.container.Logs(ctx)
		if err != nil {
			errCh <- errors.Wrap(err,
				errors.New("could not read container log"))
		}

		brdr := bufio.NewReader(rdr)

		for {
			line, _, err := brdr.ReadLine()
			if err == io.EOF {
				close(errCh)
				close(logCh)
				return
			}
			if err != nil {
				errCh <- errors.Wrap(err,
					errors.New("could not read container reader"))
			}

			logCh <- string(line)

		}

	}()
	return logCh, errCh
}
