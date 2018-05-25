package main

import (
	"foremandns/cmd"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

// LogLevel of the server
var LogLevel string

// Config file location
var cfgFile string

var (
	baseurl  string
	username string
	password string
)

// Root command
var rootCmd = &cobra.Command{
	Use:   "foremandns",
	Short: "foremandns is a simple DNS server for Forman hosts",
	Long: `A DNS server to get the hosts from forman server built with
                love by Zomato in Go.
                Complete documentation is available at https://github.com/Zomato/foremandns`,
}

func init() {
	rootCmd.AddCommand(cmd.VersionCmd)
	rootCmd.AddCommand(cmd.ServerCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Panic("Can't execute:", err)
		os.Exit(1)
	}
}
