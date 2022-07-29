package content

import (
	"bmoc/cmd/container"
	"bmoc/cmd/services"
	"bmoc/cmd/utils"
	"context"
	"fmt"
	"log"

	"github.com/dstotijn/go-notion"
	"github.com/gosimple/slug"
	"github.com/spf13/cobra"
)

var MigrateCommand = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate content from Notion to WordPress",
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("hell oworld!")

		pages, err := services.ListDraftArticles()
		if err != nil {
			log.Fatal(err)
		}

		opts := utils.Options{
			Choices:  []utils.Choice{},
			Callback: MoveArticleToWordPress,
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

func init() {

}

type WordPressPageDTO struct {
	HTML           string
	Excerpt        string
	ImagesToUpload []WordPressMediaDTO
}

type WordPressMediaDTO struct {
	Name        string
	OriginalUrl string
	Alt         string
}

func MoveArticleToWordPress(pageId string) {
	dto := WordPressPageDTO{
		ImagesToUpload: []WordPressMediaDTO{},
	}

	// Get the page from Notion
	client := container.GetNotionClient()
	// TODO: Need to grab the excerpt from here
	// page, err := client.FindPageByID(context.TODO(), pageId)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	blocks, err := client.FindBlockChildrenByID(context.TODO(), pageId, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Iterate over each block and create html
	for _, el := range blocks.Results {
		if el.Type == notion.BlockTypeHeading1 {
			dto.HTML += fmt.Sprintf("<h1>%v</h1>\n", el.Heading1.Text[0].PlainText)
		}

		if el.Type == notion.BlockTypeHeading2 {
			dto.HTML += fmt.Sprintf("<h2>%v</h2>\n", el.Heading2.Text[0].PlainText)
		}

		if el.Type == notion.BlockTypeImage {
			imgdto := WordPressMediaDTO{
				OriginalUrl: el.Image.File.URL,
			}
			if len(el.Image.Caption) > 0 {
				imgdto.Name = el.Image.Caption[0].PlainText
				imgdto.Alt = slug.Make(imgdto.Name)
			}
			// TODO: swap out original URL, implement captions
			dto.HTML += fmt.Sprintf("<img src=\"%v\" alt=\"%v\" />\n", imgdto.OriginalUrl, imgdto.Alt)
			dto.ImagesToUpload = append(dto.ImagesToUpload, imgdto)
			continue
		}

		if el.Type == notion.BlockTypeParagraph && len(el.Paragraph.Text) > 0 {
			dto.HTML += "<p>"
			for _, ptext := range el.Paragraph.Text {
				if ptext.Annotations.Bold {
					dto.HTML += fmt.Sprintf("<b>%v</b>", ptext.PlainText)
					continue
				}

				if ptext.Annotations.Italic {
					dto.HTML += fmt.Sprintf("<i>%v</i>", ptext.PlainText)
					continue
				}

				if ptext.Annotations.Code {
					// TODO: figure this one out
					continue
				}

				dto.HTML += ptext.PlainText
			}
			dto.HTML += "</p>\n"
		}

	}

	wpclient := container.GetWordPressClient()
	uplurl, err := wpclient.UploadMediaFromUrl(dto.ImagesToUpload[0].OriginalUrl, dto.ImagesToUpload[0].Name, dto.ImagesToUpload[0].Name)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(*uplurl)
}
