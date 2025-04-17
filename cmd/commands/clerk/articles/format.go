package articles

import (
	"bmoc/cmd/utils"
	"log"

	"github.com/spf13/cobra"
)

var (
	formatDocFlag string
	noCleanupFlag bool
)

var FormatCommand = &cobra.Command{
	Use:   "format",
	Short: "Formats an article for the review process.",
	Run: func(cmd *cobra.Command, args []string) {
		if formatDocFlag == "" {
			log.Println("Missing parameter, 'in (i)' is required")
			return
		}
		imgPathPrefix := "/images/blog/content"
		utils.NotionExportToMarkdown(formatDocFlag, nil, imgPathPrefix, noCleanupFlag)
	},
}

func init() {
	FormatCommand.Flags().StringVarP(&formatDocFlag, "in", "i", "", "Extracts a Notion exported zip & processes for the docs site.")
	FormatCommand.Flags().BoolVar(&noCleanupFlag, "no-cleanup", false, "If set, the original & generated temp files will not be deleted.")
}
