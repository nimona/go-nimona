package cmd

import (
	"encoding/json"

	"github.com/spf13/cobra"

	"nimona.io/go/codec"
	"nimona.io/go/primitives"
)

// blockListCmd represents the blockList command
var blockListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all known blocks",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := restClient.R().Get("/blocks")
		if err != nil {
			return err
		}

		body := resp.Body()
		blocks := []*primitives.Block{}
		if err := codec.Unmarshal(body, &blocks); err != nil {
			return err
		}

		if returnRaw {
			bs, err := json.MarshalIndent(blocks, "", "  ")
			if err != nil {
				return err
			}

			cmd.Println(string(bs))
			return nil
		}

		for _, block := range blocks {
			cmd.Println("block:")
			cmd.Println("  id:", block.ID())
			cmd.Println("  type:", block.Type)
			cmd.Println("  payload:", block.Payload)
			cmd.Println("  signer:", block.Signature.Key.Thumbprint())
			cmd.Println("")
		}
		return nil
	},
}

func init() {
	blockCmd.AddCommand(blockListCmd)
}
