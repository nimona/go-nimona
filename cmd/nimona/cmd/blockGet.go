package cmd

import (
	"encoding/json"

	"github.com/spf13/cobra"

	"nimona.io/go/encoding"
)

// blockGetCmd represents the blockGet command
var blockGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a block by its ID",
	Long:  "",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := restClient.R().Get("/blocks/" + args[0])
		if err != nil {
			return err
		}

		body := resp.Body()
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
			return nil
		}

		cmd.Println("block:")
		cmd.Println("  _id:", o.HashBase58())
		for k, v := range o.Map() {
			cmd.Printf("  %s: %v\n", k, v)
		}
		cmd.Println("")
		return nil
	},
}

func init() {
	blockCmd.AddCommand(blockGetCmd)
}
