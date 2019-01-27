package cmd

import (
	"github.com/spf13/cobra"
)

// blockListCmd represents the blockList command
var blockListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all known blocks",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		// resp, err := restClient.R().Get("/blocks")
		// if err != nil {
		// 	return err
		// }

		// body := resp.Body()
		// blocks := []map[string]interface{}{}
		// if err := object.UnmarshalInto(body, &blocks); err != nil {
		// 	return err
		// }

		// 	bs, err := json.MarshalIndent(blocks, "", "  ")
		// 	if err != nil {
		// 		return err
		// 	}
		// if viper.GetBool("raw") {

		// 	cmd.Println(string(bs))
		// 	return nil
		// }

		// for _, block := range blocks {
		// 	cmd.Println("block:")
		// 	cmd.Println("  _id:", crypto.ID(block))
		// 	for k, v := range block {
		// 		cmd.Printf("  %s: %v\n", k, v)
		// 	}
		// 	cmd.Println("")
		// }
		return nil
	},
}

func init() {
	blockCmd.AddCommand(blockListCmd)
}
