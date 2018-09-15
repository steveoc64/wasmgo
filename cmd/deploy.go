package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/steveoc64/wasmgo/cmd/deployer"
)

func init() {
	deployCmd.PersistentFlags().StringVarP(&global.Template, "template", "t", "{{ .Page }}", "Template defining the output returned by the deploy command. Variables: Page, Script, Loader, Binary.")
	deployCmd.PersistentFlags().BoolVarP(&global.Json, "json", "j", false, "Return all template variables as a json blob from the deploy command.")
	rootCmd.AddCommand(deployCmd)
}

var deployCmd = &cobra.Command{
	Use:   "deploy [package]",
	Short: "Compile and deploy",
	Long:  "Compiles Go to WASM and deploys to the jsgo.io CDN.",
	Args:  cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			global.Path = args[0]
		}
		d, err := deployer.New(global)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
			os.Exit(1)
		}
		if err := d.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
			os.Exit(1)
		}
	},
}
