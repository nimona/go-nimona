package cmd

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"nimona.io/pkg/object"
)

// objectListenCmd represents the objectListen command
var objectListenCmd = &cobra.Command{
	Use:   "listen",
	Short: "Listen for new incoming objects matching a pattern",
	Long:  "",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		url := viper.GetString("api") + "/streams/" + args[0]
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

			o, err := object.FromBytes(body)
			if err != nil {
				return err
			}

			if viper.GetBool("raw") {
				bs, err := json.MarshalIndent(o.ToMap(), "", "  ")
				if err != nil {
					return err
				}

				cmd.Println(string(bs))
				continue
			}

			cmd.Println("object:")
			cmd.Println("  _id:", o.HashBase58())
			for k, v := range o.ToMap() {
				cmd.Printf("  %s: %v\n", k, v)
			}
			cmd.Println("")
		}
	},
}

func init() {
	objectCmd.AddCommand(objectListenCmd)
}
