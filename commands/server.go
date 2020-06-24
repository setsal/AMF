package commands

import (
	"log"

	"github.com/setsal/go-AMF/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start amf HTTP server",
	Long:  `Start amf HTTP server`,
	RunE:  executeServer,
}

func init() {
	RootCmd.AddCommand(serverCmd)

	serverCmd.PersistentFlags().String("bind", "", "HTTP server bind to IP address (default: loop back interface")
	viper.BindPFlag("bind", serverCmd.PersistentFlags().Lookup("bind"))
	viper.SetDefault("bind", "127.0.0.1")

	serverCmd.PersistentFlags().String("port", "", "HTTP server port number (default: 3000")
	viper.BindPFlag("port", serverCmd.PersistentFlags().Lookup("port"))
	viper.SetDefault("port", "3000")

}

func executeServer(cmd *cobra.Command, args []string) (err error) {
	logDir := viper.GetString("log-dir")
	log.Println("Starting HTTP Server...")

	if err = server.Start(logDir); err != nil {
		log.Println("Failed to start server.", err)
		return err
	}

	return nil
}
