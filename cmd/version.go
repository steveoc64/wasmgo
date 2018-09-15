package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/steveoc64/wasmgo/cmd/deployer"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show client version number",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(deployer.CLIENT_VERSION)
	},
}
