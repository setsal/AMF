package commands

import (
	"fmt"
	"os"

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

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $PROJECT/.go-AMF.yaml)")
	RootCmd.PersistentFlags().String("log-dir", "", "/path/to/log")

	viper.BindPFlag("log-dir", RootCmd.PersistentFlags().Lookup("log-dir"))
	viper.SetDefault("log-dir", util.GetDefaultLogDir())
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {

	// add config path in current project dir
	viper.SetConfigType("yaml")

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName(".go-AMF")
		viper.AddConfigPath(os.Getenv("PWD"))
	}

	err := viper.ReadInConfig()
	if err != nil { // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s", err))
	}

	viper.AutomaticEnv() // read in environment variables that match
}
