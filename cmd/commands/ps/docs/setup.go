package docs

import (
	"bmoc/cmd/services"
	"bmoc/cmd/utils"
	"log"

	"github.com/dstotijn/go-notion"
	"github.com/spf13/cobra"
)

// PsCmd represents the ps command
var SetupDocCmd = &cobra.Command{
	Use:   "setup",
	Short: "Sets up the project bits for PlanetScale docs in Notion",
	Run:   run,
}

func run(cmd *cobra.Command, args []string) {
	/* List PlanetScale docs in Notion
	- filters: Area = PlanetScale, Status = Selected
	*/

	/* For selected:
	- set icon to Area icon
	- create project
	- link project to content item
	- create tasks for project
		- Write draft
		- Self review
			- adheres to content guide
			- run through grammarly
			- all images have captions
			- GetHelp is added
		- Migrate to GitHub
			- export to zip & process w/bmoc
			- create PR
			- tag holly & legal
	- Cleanup old resources
	- Notify Jenn
	*/

	// Get content items to present
	area := utils.GetConfigString("NOTION_AREA_PLANETSCALE", true)
	status := "Selected"
	pages, err := services.ListContentItems(&status, &area)
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
}

func callback(contentItemId string) {
	services.SetupDoc(contentItemId)
}
