package articles

import (
	"bmoc/cmd/services"
	"bmoc/cmd/utils"
	"log"

	"github.com/dstotijn/go-notion"
	"github.com/spf13/cobra"
)

var SetupCommand = &cobra.Command{
	Use: "setup",
	Aliases: []string{
		"s",
	},
	Short: "Sets up the project in Notion for a selected article",
	Run: func(cmd *cobra.Command, args []string) {

		area := utils.GetConfigString("NOTION_AREA_PLANETSCALE", true)
		status := "Selected"
		pages, err := services.ListContentItems(&status, &area)
		if err != nil {
			log.Fatal(err)
		}

		opts := utils.Options{
			Title:    "Setup article project:",
			Choices:  []utils.Choice{},
			Callback: setupCommandCallback,
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

func setupCommandCallback(contentItemId string) {
	services.SetupContentProject(contentItemId)
}
