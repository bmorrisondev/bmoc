package articles

import (
	"github.com/spf13/cobra"
)

var ArticlesCommand = &cobra.Command{
	Use: "articles",
	Aliases: []string{
		"a",
	},
	Short: "Automations on drafting and publishing articles for the PlanetScale blog",
}

func init() {
	ArticlesCommand.AddCommand(SetupCommand)
}
