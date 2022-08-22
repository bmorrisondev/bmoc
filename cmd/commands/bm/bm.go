package bm

import (
	"bmoc/cmd/commands/bm/content"

	"github.com/spf13/cobra"
)

var BmCommand = &cobra.Command{
	Use:   "bm",
	Short: "Commands around personal content & automation",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

func init() {
	BmCommand.AddCommand(content.ContentCommand)
}
