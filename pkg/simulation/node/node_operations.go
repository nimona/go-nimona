package node

import (
	"bufio"
	"io"

	"github.com/tidwall/gjson"
	"nimona.io/pkg/context"
	"nimona.io/pkg/errors"
)

// Logs filters the node json logs for the specific tag
func (n *Node) Logs(tag string) ([]string, error) {
	ctx := context.Background()

	rdr, err := n.container.Logs(ctx)
	if err != nil {
		return nil, errors.Wrap(err,
			errors.New("could not read container log"))
	}

	brdr := bufio.NewReader(rdr)
	logs := []string{}

	for {
		line, _, err := brdr.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, errors.Wrap(err,
				errors.New("could not read container reader"))
		}

		value := gjson.Get(string(line), tag)
		logs = append(logs, value.String())
	}

	return logs, nil
}
