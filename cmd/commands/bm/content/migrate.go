package content

import (
	"bmoc/cmd/container"
	"bmoc/cmd/services"
	"bmoc/cmd/utils"
	"fmt"
	"log"
	"strings"

	"github.com/dstotijn/go-notion"
	"github.com/gosimple/slug"
	"github.com/spf13/cobra"
)

var MigrateCommand = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate content from Notion to WordPress",
	Run: func(cmd *cobra.Command, args []string) {
		pages, err := services.ListDraftArticles()
		if err != nil {
			log.Fatal(err)
		}

		opts := utils.Options{
			Choices:  []utils.Choice{},
			Callback: migrateCommandCallback,
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

func migrateCommandCallback(pageId string) {
	// Get DTO
	dto := services.NotionToWordPressPage(pageId)

	// Send images to WordPress & update html
	wpclient := container.GetWordPressClient()
	for _, el := range dto.ImagesToUpload {
		if el.Slug == "" {
			continue
		}
		// TODO: Need to get the id of the returned image so I can link them in the post
		uplurl, err := wpclient.UploadMediaFromUrl(el.OriginalUrl, el.Name, el.Name)
		if err != nil {
			log.Fatal(err)
		}
		replaceWith := fmt.Sprintf(`<figure class="wp-block-image size-full"><img loading="lazy" src="%v" alt="%v"><figcaption>%v</figcaption></figure>`, *uplurl, el.Name, el.Name)
		dto.HTML = strings.Replace(dto.HTML, el.Tag, replaceWith, 1)
	}
	log.Println(dto.HTML)

	// Send post to WordPress
	req := services.WPPagePostRequest{
		Date:     "2022-08-13T00:00:00",
		Slug:     slug.Make(dto.Title),
		Status:   "publish",
		Title:    dto.Title,
		Content:  dto.HTML,
		AuthorId: 2,
		Excerpt:  dto.Excerpt,
	}
	log.Println(req)
	_, err := wpclient.CreatePost(req)
	if err != nil {
		log.Fatal(err)
	}
	// TODO: better message here, maybe open it in wp right away
	log.Println("Done!")
}
