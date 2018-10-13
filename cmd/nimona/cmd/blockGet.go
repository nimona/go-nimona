package cmd

import (
	"encoding/json"

	"github.com/spf13/cobra"

	"nimona.io/go/codec"
	"nimona.io/go/primitives"
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
		block := &primitives.Block{}
		if err := codec.Unmarshal(body, block); err != nil {
			return err
		}

		if returnRaw {
			bs, err := json.MarshalIndent(block, "", "  ")
			if err != nil {
				return err
			}

			cmd.Println(string(bs))
			return nil
		}

		cmd.Println("block:")
		cmd.Println("  id:", block.ID())
		cmd.Println("  type:", block.Type)
		cmd.Println("  payload:", block.Payload)
		cmd.Println("  signer:", block.Signature.Key.Thumbprint())
		cmd.Println("")
		return nil
	},
}

func init() {
	blockCmd.AddCommand(blockGetCmd)
}
