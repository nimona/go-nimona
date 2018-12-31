package cmd

import (
	"context"
	"errors"

	"github.com/spf13/cobra"
	"nimona.io/cmd/nimona/cmd/providers"
)

var (
	platform       string
	dockerTag      string
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

// provisionCmd represents the daemon command
var provisionCmd = &cobra.Command{
	Use:   "provision",
	Short: "Provision a daemon peer on a remote provider",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		switch platform {
		case "do":
			dop, err := providers.NewDigitalocean(token)
			if err != nil {
				return err
			}

			cmd.Printf("Starting server: %s\n", hostname)

			ip, err := dop.NewInstance(dockerTag, hostname, sshFingerprint, size, region)
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
	rootCmd.AddCommand(provisionCmd)

	provisionCmd.PersistentFlags().StringVar(
		&platform,
		"platform",
		"do",
		"target platform",
	)
	provisionCmd.PersistentFlags().StringVar(
		&hostname,
		"hostname",
		"",
		"peer hostname",
	)
	provisionCmd.PersistentFlags().StringVar(
		&dockerTag,
		"docker-tag",
		"",
		"docker tag",
	)
	provisionCmd.PersistentFlags().StringVar(
		&token,
		"token",
		"",
		"platform access token",
	)
	provisionCmd.PersistentFlags().StringVar(
		&sshFingerprint,
		"ssh-fingerprint",
		"",
		"ssh fingerprint",
	)
	provisionCmd.PersistentFlags().StringVar(
		&size,
		"size",
		"",
		"instance size",
	)
	provisionCmd.PersistentFlags().StringVar(
		&region,
		"region",
		"",
		"instance region",
	)
}
