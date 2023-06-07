package docs

import (
	"bmoc/cmd/utils"
	"log"

	"github.com/spf13/cobra"
)

var (
	formatDocFlag string
	noCleanupFlag bool
)

// PsCmd represents the ps command
var FormatCmd = &cobra.Command{
	Use:   "format",
	Short: "",
	Run: func(cmd *cobra.Command, args []string) {
		if formatDocFlag == "" {
			log.Println("Missing parameter, 'in (i)' is required")
			return
		}
		imgPathPrefix := "/assets/docs"
		utils.NotionExportToMarkdown(formatDocFlag, nil, imgPathPrefix, noCleanupFlag)
	},
}

func init() {
	FormatCmd.Flags().StringVarP(&formatDocFlag, "in", "i", "", "Extracts a Notion exported zip & processes for the docs site.")
	FormatCmd.Flags().BoolVar(&noCleanupFlag, "no-cleanup", false, "If set, the original & generated temp files will not be deleted.")
}
