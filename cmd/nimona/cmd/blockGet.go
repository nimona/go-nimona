package cmd

import (
	"encoding/json"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"nimona.io/pkg/object"
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
			return nil
		}

		cmd.Println("block:")
		cmd.Println("  _id:", o.HashBase58())
		for k, v := range o.ToMap() {
			cmd.Printf("  %s: %v\n", k, v)
		}
		cmd.Println("")
		return nil
	},
}

func init() {
	blockCmd.AddCommand(blockGetCmd)
}
