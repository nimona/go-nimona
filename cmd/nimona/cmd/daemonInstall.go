package cmd

import (
	"context"
	"errors"

	"github.com/spf13/cobra"
	"nimona.io/go/cmd/nimona/cmd/providers"
)

var (
	platform       string
	hostname       string
	token          string
	sshFingerprint string
	size           string
	region         string
)

var (
	// ErrNoPlatform returned when no platform has been provided
	ErrNoPlatform = errors.New("missing platform")
)

// daemonInstallCmd represents the daemon command
var daemonInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install a peer as a daemon in a remote provider",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		switch platform {
		case "do":
			dop, err := providers.NewDigitalocean(token)
			if err != nil {
				return err
			}

			cmd.Printf("Starting server: %s\n", hostname)

			ip, err := dop.NewInstance(hostname, sshFingerprint, size, region)
			if err != nil {
				return err
			}

			cmd.Printf("Server created. IP: %s\n", ip)

			// If hostname specified create domain entry
			if hostname == "" {
				return nil
			}

			cmd.Printf("Updating domain: %s with ip: %s\n", hostname, ip)

			err = dop.UpdateDomain(context.Background(),
				hostname, ip)
			if err != nil {
				return err
			}
			cmd.Println("Domain updated")

		case "":
			return ErrNoPlatform
		}

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
		"ssh-fingerprint",
		"",
		"ssh fingerprint",
	)
	daemonInstallCmd.PersistentFlags().StringVar(
		&size,
		"size",
		"",
		"instance size",
	)
	daemonInstallCmd.PersistentFlags().StringVar(
		&region,
		"region",
		"",
		"instance region",
	)
}
