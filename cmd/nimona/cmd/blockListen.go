package cmd

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"

	"nimona.io/go/encoding"
)

// blockListenCmd represents the blockListen command
var blockListenCmd = &cobra.Command{
	Use:   "listen",
	Short: "Listen for new incoming blocks matching a pattern",
	Long:  "",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		url := apiAddress + "/streams/" + args[0]
		url = strings.Replace(url, "http", "ws", 1)
		dialer := websocket.DefaultDialer
		headers := http.Header{}
		headers.Set("Content-Type", "application/cbor")
		c, _, err := dialer.Dial(url, headers)
		if err != nil {
			return err
		}

		defer c.Close()

		for {
			wsMsgType, body, err := c.ReadMessage()
			if err != nil {
				return err
			}

			if wsMsgType != 2 {
				continue
			}

			o, err := encoding.NewObjectFromBytes(body)
			if err != nil {
				return err
			}

			if returnRaw {
				bs, err := json.MarshalIndent(o.Map(), "", "  ")
				if err != nil {
					return err
				}

				cmd.Println(string(bs))
				continue
			}

			cmd.Println("block:")
			cmd.Println("  _id:", o.HashBase58())
			for k, v := range o.Map() {
				cmd.Printf("  %s: %v\n", k, v)
			}
			cmd.Println("")
		}
	},
}

func init() {
	blockCmd.AddCommand(blockListenCmd)
}
