package cmd

import (
	"github.com/spf13/cobra"
)

// objectListCmd represents the objectList command
var objectListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all known objects",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		// resp, err := restClient.R().Get("/objects")
		// if err != nil {
		// 	return err
		// }

		// body := resp.Body()
		// objects := []map[string]interface{}{}
		// if err := object.UnmarshalInto(body, &objects); err != nil {
		// 	return err
		// }

		// 	bs, err := json.MarshalIndent(objects, "", "  ")
		// 	if err != nil {
		// 		return err
		// 	}
		// if viper.GetBool("raw") {

		// 	cmd.Println(string(bs))
		// 	return nil
		// }

		// for _, object := range objects {
		// 	cmd.Println("object:")
		// 	cmd.Println("  _id:", crypto.ID(object))
		// 	for k, v := range object {
		// 		cmd.Printf("  %s: %v\n", k, v)
		// 	}
		// 	cmd.Println("")
		// }
		return nil
	},
}

func init() {
	objectCmd.AddCommand(objectListCmd)
}
