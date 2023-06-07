package docs

import (
	"github.com/spf13/cobra"
)

// PsCmd represents the ps command
var DocsCommand = &cobra.Command{
	Use:   "docs",
	Short: "Functions for automating stuff around the docs site",
}

func init() {
	DocsCommand.AddCommand(FormatCmd)
	DocsCommand.AddCommand(SetupDocCmd)
	DocsCommand.AddCommand(ExportCommand)
}
