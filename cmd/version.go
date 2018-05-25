package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var (
	// Version string
	Version string
	// BuildDate string
	BuildDate string
)

// VersionCmd for check the version of the server
var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of foremandns",
	Long:  `All software has versions. This is foremandns's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(fmt.Sprintf("foremandns %s Build date %s", Version, BuildDate))
	},
}
