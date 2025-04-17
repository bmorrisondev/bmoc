package clerk

import (
	"bmoc/cmd/commands/clerk/articles"

	"github.com/spf13/cobra"
)

// PsCmd represents the ps command
var ClerkCmd = &cobra.Command{
	Use:   "c",
	Short: "Clerk stuffz",
}

func init() {
	// Subcommands
	// ClerkCmd.AddCommand(docs.DocsCommand)
	ClerkCmd.AddCommand(articles.ArticlesCommand)

	// Flags
}
