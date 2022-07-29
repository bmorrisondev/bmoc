package config

import (
	"fmt"

	"github.com/spf13/cobra"
)

// configCmd represents the config command
var ConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Modify configuration values.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("config called")
	},
}

func init() {
	ConfigCmd.AddCommand(SetCmd)
	ConfigCmd.AddCommand(ShowCmd)
	ConfigCmd.AddCommand(DeleteCmd)
}
