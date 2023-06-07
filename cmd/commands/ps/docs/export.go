package docs

import (
	"bmoc/cmd/services"
	"bmoc/cmd/services/ntomd"
	"bmoc/cmd/utils"
	"log"
	"os"

	"github.com/dstotijn/go-notion"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var ExportCommand = &cobra.Command{
	Use:   "export",
	Short: "Exports a doc from Notion to Markdown.",
	Run: func(cmd *cobra.Command, args []string) {
		// imgPathPrefix := "/assets/docs"

		// Get content items to present
		area := utils.GetConfigString("NOTION_AREA_PLANETSCALE", true)
		status := "Ready"
		contentType := "Docs"
		pages, err := services.ListContentItems(&status, &area, &contentType)
		if err != nil {
			log.Fatal(err)
		}

		opts := utils.Options{
			Choices:  []utils.Choice{},
			Callback: exportCmdCallback,
		}

		for _, el := range pages {
			c := utils.Choice{Id: el.ID}
			props := el.Properties.(notion.DatabasePageProperties)
			c.Name = props["Name"].Title[0].PlainText
			opts.Choices = append(opts.Choices, c)
		}

		utils.PresentSelector(opts)

	},
}

func exportCmdCallback(contentItemId string) {
	notionKey := viper.GetString("NOTION_KEY")
	log.Println(notionKey)
	md, err := ntomd.GetMarkdownStringFromNotionPage(&notionKey, &contentItemId)
	if err != nil {
		log.Fatal(err)
	}

	os.WriteFile("./test.md", []byte(*md), 0664)
}
