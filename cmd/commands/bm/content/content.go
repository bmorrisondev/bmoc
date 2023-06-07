package content

import (
	"github.com/spf13/cobra"
)

var ContentCommand = &cobra.Command{
	Use: "content",
	Aliases: []string{
		"c",
	},
	Short: "Automate all the contents!",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func init() {
	ContentCommand.AddCommand(SetupCommand)
}
