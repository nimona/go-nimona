package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/user"
	"path"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/profile"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	resty "gopkg.in/resty.v1"
)

var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"

	env        string
	dataDir    string
	cfgFile    string
	apiAddress string
	apiToken   string
	returnRaw  bool

	restClient *resty.Client

	config = &Config{}
)

var rootCmd = &cobra.Command{
	Use:   "nimona",
	Short: "",
	Long:  "",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		restClient = resty.New().
			SetHostURL(viper.GetString("api")).
			SetTimeout(10*time.Second).
			SetHeader("Content-Type", "application/cbor").
			SetHeader("Authorization", viper.GetString("api_token")).
			SetContentLength(true).
			SetRESTMode().
			SetRedirectPolicy(resty.FlexibleRedirectPolicy(5))

		if strings.ToLower(viper.GetString("env")) == "dev" {
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
	viper.BindPFlag("api", rootCmd.PersistentFlags().Lookup("api"))

	rootCmd.PersistentFlags().StringVar(
		&apiToken,
		"api-token",
		"",
		"api token",
	)
	viper.BindPFlag("api_token", rootCmd.PersistentFlags().Lookup("api-token"))

	rootCmd.PersistentFlags().StringVarP(
		&env,
		"env",
		"e",
		"PROD",
		"environment; used for debugging",
	)
	viper.BindPFlag("env", rootCmd.PersistentFlags().Lookup("env"))

	rootCmd.PersistentFlags().BoolVar(
		&returnRaw,
		"raw",
		false,
		"return raw response",
	)
	viper.BindPFlag("raw", rootCmd.PersistentFlags().Lookup("raw"))

	rootCmd.PersistentFlags().StringVar(
		&dataDir,
		"data-dir",
		"",
		"data directory",
	)
	_ = viper.BindPFlag(
		"data_dir",
		rootCmd.PersistentFlags().Lookup("data-dir"),
	)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if dataDir == "" {
		usr, err := user.Current()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		dataDir = path.Join(usr.HomeDir, ".nimona")
	}

	if cfgFile == "" {
		cfgFile = path.Join(dataDir, "config.json")
	}

	viper.SetConfigFile(cfgFile)
	viper.SetConfigType("json")
	viper.SetEnvPrefix("NIMONA")
	viper.AutomaticEnv() // read in environment variables that match
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
		jsonFile, err := os.Open(cfgFile)
		if err != nil {
			log.Fatal("could not read config")
		}
		defer jsonFile.Close() // nolint
		jsonBytes, _ := ioutil.ReadAll(jsonFile)
		// NOTE(geoah): do not use viper.Unmarshal as it doesn't obey the json
		// tags, but instead requires the mapstructure tag.
		if err := json.Unmarshal(jsonBytes, config); err != nil {
			log.Fatal("could not unmarshal config", err)
		}
	}
}

func Root() *cobra.Command {
	return rootCmd
}
