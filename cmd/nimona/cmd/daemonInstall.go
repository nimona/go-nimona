package cmd

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"nimona.io/go/cmd/nimona/cmd/providers"
)

var (
	platform       string
	hostname       string
	token          string
	sshFingerprint string
)

// daemonInstallCmd represents the daemon command
var daemonInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install a peer as a daemon in a remote provider",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		if platform == "" {
			return errors.New("no platform defined")
		}
		if token == "" {
			return errors.New("missing platform token")
		}

		dop := providers.NewDigitalocean(token)

		dop.NewInstance("test")

		return nil
	},
}

func init() {
	daemon.AddCommand(daemonInstallCmd)

	daemonInstallCmd.PersistentFlags().StringVar(
		&platform,
		"platform",
		"",
		"target platform",
	)
	daemonInstallCmd.PersistentFlags().StringVar(
		&hostname,
		"hostname",
		"",
		"peer hostname",
	)
	daemonInstallCmd.PersistentFlags().StringVar(
		&token,
		"token",
		"",
		"platform access token",
	)
	daemonInstallCmd.PersistentFlags().StringVar(
		&sshFingerprint,
		"sshFingerprint",
		"",
		"sshFingerprint",
	)
}
