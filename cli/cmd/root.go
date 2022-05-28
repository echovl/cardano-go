package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfg Config
)

type Config struct {
	BlockfrostProjectID string `mapstructure:"blockfrost_project_id"`
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "cardano-wallet",
	Short: "A CLI application to manage wallets in Cardano.",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	viper.SetConfigName("cwallet")
	viper.SetConfigType("yaml")

	xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfigHome == "" {
		viper.AddConfigPath("$HOME/.config")
	} else {
		viper.AddConfigPath("$XDG_CONFIG_HOME")
	}

	if err := viper.ReadInConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "parsing config file: %s\n", err)
		os.Exit(1)
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		fmt.Fprintf(os.Stderr, "parsing config file: %s\n", err)
		os.Exit(1)
	}
}
