package cmd

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/pkg/profile"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	resty "gopkg.in/resty.v1"
)

var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"

	env              string
	cfgFile          string
	apiAddress       string
	apiToken         string
	announceHostname string
	returnRaw        bool

	restClient *resty.Client
)

var rootCmd = &cobra.Command{
	Use:   "nimona",
	Short: "",
	Long:  "",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		restClient = resty.New().
			SetHostURL(apiAddress).
			SetTimeout(10*time.Second).
			SetHeader("Content-Type", "application/cbor").
			SetHeader("Authorization", apiToken).
			SetContentLength(true).
			SetRESTMode().
			SetRedirectPolicy(resty.FlexibleRedirectPolicy(5))

		if strings.ToLower(env) == "dev" {
			fmt.Println("Running in development mode, this will be very verbose")
			defer profile.Start(profile.MemProfile).Stop()
			go http.ListenAndServe(":1234", nil)
			gin.SetMode(gin.DebugMode)
		} else {
			gin.SetMode(gin.ReleaseMode)
		}

		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(
		&cfgFile,
		"config",
		"c",
		"",
		"config file  (default is .nimona.yaml or .nimona.json in $HOME)",
	)

	rootCmd.PersistentFlags().StringVar(
		&apiAddress,
		"api",
		"http://localhost:8030/api/v1",
		"api address",
	)

	rootCmd.PersistentFlags().StringVar(
		&apiToken,
		"api-token",
		"",
		"api token",
	)

	rootCmd.PersistentFlags().StringVar(
		&announceHostname,
		"announce-hostname",
		"",
		"set and announce local dns address",
	)

	rootCmd.PersistentFlags().StringVar(
		&env,
		"env",
		"PROD",
		"environment; used for debugging",
	)

	rootCmd.PersistentFlags().BoolVar(
		&returnRaw,
		"raw",
		false,
		"return raw response",
	)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".nimona" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".nimona")
	}

	viper.SetEnvPrefix("NIMONA")
	viper.AutomaticEnv() // read in environment variables that match
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func Root() *cobra.Command {
	return rootCmd
}
