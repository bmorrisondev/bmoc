package ps

import (
	"bmoc/cmd/commands/ps/articles"
	"bmoc/cmd/commands/ps/docs"

	"github.com/spf13/cobra"
)

// PsCmd represents the ps command
var PsCmd = &cobra.Command{
	Use:   "ps",
	Short: "PlanetScale stuffz",
}

func init() {
	// Subcommands
	PsCmd.AddCommand(docs.DocsCommand)
	PsCmd.AddCommand(articles.ArticlesCommand)

	// Flags
}
