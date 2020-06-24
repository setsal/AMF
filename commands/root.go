package commands

import (
	"fmt"

	"github.com/setsal/go-AMF/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:           "go-AMF",
	Short:         "Automatic message forwarding",
	Long:          `Automatic message forwarding`,
	SilenceErrors: true,
	SilenceUsage:  true,
}

func init() {
	cobra.OnInitialize(initConfig)

	// 暫時不知這行用意
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.go-amf.yaml)")
	RootCmd.PersistentFlags().String("log-dir", "", "/path/to/log")

	viper.BindPFlag("log-dir", RootCmd.PersistentFlags().Lookup("log-dir"))
	viper.SetDefault("log-dir", util.GetDefaultLogDir())
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName(".go-cvemap")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
