package node

import (
	"bufio"
	"fmt"
	"io"

	"nimona.io/pkg/context"
)

// Logs filters the node json logs for the specific tag
func (n *Node) Logs() (chan string, chan error) {
	logCh := make(chan string)
	errCh := make(chan error)

	go func() {
		ctx := context.Background()

		rdr, err := n.container.Logs(ctx)
		if err != nil {
			errCh <- fmt.Errorf("could not read container log: %w", err)
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
				errCh <- fmt.Errorf("could not read container reader: %w", err)
			}

			logCh <- string(line)

		}
	}()
	return logCh, errCh
}
