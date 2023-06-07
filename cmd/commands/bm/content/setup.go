package content

import (
	"bmoc/cmd/services"
	"bmoc/cmd/utils"
	"log"

	"github.com/dstotijn/go-notion"
	"github.com/spf13/cobra"
)

type ProjectTask map[string][]string

var SetupCommand = &cobra.Command{
	Use:   "setup",
	Short: "Sets up a content item for management.",
	Run: func(cmd *cobra.Command, args []string) {

		area := utils.GetConfigString("NOTION_AREA_CREATOR", true)
		status := "Selected"
		pages, err := services.ListContentItems(&status, &area, nil)
		if err != nil {
			log.Fatal(err)
		}

		opts := utils.Options{
			Choices:  []utils.Choice{},
			Callback: callback,
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

func callback(contentItemId string) {
	services.SetupContentProject(contentItemId)
}
