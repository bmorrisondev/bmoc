package services

import (
	"bmoc/cmd/models"
	"bmoc/cmd/utils"
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/dstotijn/go-notion"
	"github.com/gosimple/slug"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

var (
	client *notion.Client
)

func setup() {
	notionKey := viper.GetString("NOTION_KEY")
	if notionKey == "" {
		log.Fatal("ERROR: Config 'NOTION_KEY' is required to be set.")
	}
	client = notion.NewClient(notionKey)
}

func GetHeaderBlockText(blockUrl string) *string {
	setup()
	spl := strings.Split(blockUrl, "-")
	blockId := spl[len(spl)-1]
	bl, err := client.FindBlockByID(context.TODO(), blockId)
	if err != nil {
		log.Println(fmt.Sprintf("WARN: Unable to parse title of block: %v // url: %v", blockId, blockUrl))
		return nil
	}

	return &bl.ChildPage.Title
}

func ListDraftArticles() ([]notion.Page, error) {
	setup()
	dbid := viper.GetString("NOTION_CONTENT_DB")
	if dbid == "" {
		log.Fatal("ERROR: Config 'NOTION_CONTENT_DB' is required to be set.")
	}

	filter := notion.DatabaseQuery{
		Filter: &notion.DatabaseQueryFilter{
			Property: "Status",
			Select: &notion.SelectDatabaseQueryFilter{
				Equals: "Draft In Progress",
			},
		},
	}

	res, err := client.QueryDatabase(context.TODO(), dbid, &filter)
	if err != nil {
		return nil, errors.Wrap(err, "(ListDraftArticles) client.QueryDatabase")
	}

	return res.Results, nil
}

func ListContentItems(status *string, area *string) ([]notion.Page, error) {
	setup()
	dbid := viper.GetString("NOTION_CONTENT_DB")
	if dbid == "" {
		log.Fatal("ERROR: Config 'NOTION_CONTENT_DB' is required to be set.")
	}

	filter := notion.DatabaseQuery{
		Filter: &notion.DatabaseQueryFilter{
			And: []notion.DatabaseQueryFilter{},
		},
	}

	if status != nil {
		filter.Filter.And = append(filter.Filter.And,
			notion.DatabaseQueryFilter{
				Property: "Status",
				Select: &notion.SelectDatabaseQueryFilter{
					Equals: *status,
				},
			},
		)
	}

	if area != nil {
		filter.Filter.And = append(filter.Filter.And,
			notion.DatabaseQueryFilter{
				Property: "Area",
				Relation: &notion.RelationDatabaseQueryFilter{
					Contains: *area,
				},
			},
		)
	}

	res, err := client.QueryDatabase(context.TODO(), dbid, &filter)
	if err != nil {
		return nil, errors.Wrap(err, "(ListDraftArticles) client.QueryDatabase")
	}

	return res.Results, nil
}

func SetupContentProject(contentItemId string) {
	setup()

	// Get the Content Item page
	page, err := client.FindPageByID(context.TODO(), contentItemId)
	if err != nil {
		log.Fatal(err)
	}
	props := page.Properties.(notion.DatabasePageProperties)
	projectName := props["Name"].Title[0].PlainText
	areaId := props["Area"].Relation[0].ID
	types := []string{}
	for _, el := range props["Type"].MultiSelect {
		types = append(types, el.Name)
	}

	// Get the Area Page
	area, err := client.FindPageByID(context.TODO(), areaId)
	if err != nil {
		log.Fatal(err)
	}
	areaProps := area.Properties.(notion.DatabasePageProperties)
	areaName := areaProps["Name"].Title[0].PlainText

	// Update page with icon & draft status
	_, err = updateContentItemPage(contentItemId)
	if err != nil {
		log.Fatal(err)
	}

	// Create project
	projectPage, err := createContentProject(projectName, contentItemId)
	log.Println("Created project:", projectPage.ID)

	// Create tasks
	projectTasks := models.BuildProjectTaskList(areaName, types)
	for task, subtasks := range projectTasks {
		_, err = createContentItemTask(projectPage.ID, contentItemId, task, subtasks)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func updateContentItemPage(contentItemId string) (notion.Page, error) {
	updParams := notion.UpdatePageParams{
		// TODO: Grab the icon from the Area and apply it here
		// Icon: &notion.Icon{
		// 	File: area.Icon.File,
		// },
		DatabasePageProperties: notion.DatabasePageProperties{
			"Status": notion.DatabasePageProperty{
				Select: &notion.SelectOptions{
					Name: "Ready",
				},
			},
		},
	}
	return client.UpdatePage(context.TODO(), contentItemId, updParams)
}

func createContentProject(projectName, contentItemId string) (notion.Page, error) {
	dbid := utils.GetConfigString("NOTION_PROJECTS_DB", true)
	params := notion.CreatePageParams{
		ParentType: notion.ParentTypeDatabase,
		ParentID:   dbid,
		DatabasePageProperties: &notion.DatabasePageProperties{
			"Name": notion.DatabasePageProperty{
				Title: []notion.RichText{
					{
						Text: &notion.Text{
							Content: projectName,
						},
					},
				},
			},
			"Status": notion.DatabasePageProperty{
				Select: &notion.SelectOptions{
					Name: "Active",
				},
			},
			"Type": notion.DatabasePageProperty{
				Select: &notion.SelectOptions{
					Name: "Content",
				},
			},
			"Content Item": notion.DatabasePageProperty{
				Relation: []notion.Relation{
					{
						ID: contentItemId,
					},
				},
			},
		},
	}

	return client.CreatePage(context.TODO(), params)
}

func createContentItemTask(projectId string, contentItemId string, task string, subtasks []string) (notion.Page, error) {
	dbid := utils.GetConfigString("NOTION_TASKS_DB", true)
	isChecked := true
	params := notion.CreatePageParams{
		ParentType: notion.ParentTypeDatabase,
		ParentID:   dbid,
		DatabasePageProperties: &notion.DatabasePageProperties{
			"Name": notion.DatabasePageProperty{
				Title: []notion.RichText{
					{
						Text: &notion.Text{
							Content: task,
						},
					},
				},
			},
			"Sprint Status": notion.DatabasePageProperty{
				Select: &notion.SelectOptions{
					Name: "To Do",
				},
			},
			"Status": notion.DatabasePageProperty{
				Select: &notion.SelectOptions{
					Name: "Next Action",
				},
			},
			"Context": notion.DatabasePageProperty{
				Select: &notion.SelectOptions{
					Name: "💻.  Computer",
				},
			},
			"Processed": notion.DatabasePageProperty{
				Checkbox: &isChecked,
			},
			"Project": notion.DatabasePageProperty{
				Relation: []notion.Relation{
					{
						ID: projectId,
					},
				},
			},
			"Content Item": notion.DatabasePageProperty{
				Relation: []notion.Relation{
					{
						ID: contentItemId,
					},
				},
			},
		},
	}

	if subtasks != nil {
		params.Children = []notion.Block{}
		for _, el := range subtasks {
			params.Children = append(params.Children, notion.Block{
				ToDo: &notion.ToDo{
					RichTextBlock: notion.RichTextBlock{
						Text: []notion.RichText{
							{
								Text: &notion.Text{
									Content: el,
								},
							},
						},
					},
				},
			})
		}
	}
	_, err := client.CreatePage(context.TODO(), params)
	if err != nil {
		log.Fatal(err)
	}

	return client.CreatePage(context.TODO(), params)

	// 	// Get current iterator by area
	// 	dbid = utils.GetConfigString("NOTION_ITERATOR_DB", true)
	// 	results, err := client.QueryDatabase(context.Background(), dbid, &notion.DatabaseQuery{
	// 		Filter: &notion.DatabaseQueryFilter{
	// 			Property: "Area",
	// 			Relation: &notion.RelationDatabaseQueryFilter{
	// 				Contains: areaId,
	// 			},
	// 		},
	// 	})
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	if len(results.Results) >= 1 {

	// 	}
	// 	iterationProps := iteration.Properties.(notion.DatabasePageProperties)
	// 	// Copy content template folder

	// }
}

func NotionToWordPressPage(pageId string) models.WordPressPageDTO {
	setup()
	dto := models.WordPressPageDTO{
		ImagesToUpload: []models.WordPressMediaDTO{},
	}

	// Get the page from Notion

	page, err := client.FindPageByID(context.TODO(), pageId)
	if err != nil {
		log.Fatal(err)
	}
	props := page.Properties.(notion.DatabasePageProperties)
	if len(props["Excerpt"].RichText) > 0 {
		dto.Excerpt = props["Excerpt"].RichText[0].PlainText
	}

	blocks, err := client.FindBlockChildrenByID(context.TODO(), pageId, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Iterate over each block and create html
	for _, el := range blocks.Results {
		if el.Type == notion.BlockTypeHeading1 {
			dto.HTML += fmt.Sprintf("<h1>%v</h1>\n", el.Heading1.Text[0].PlainText)
			continue
		}

		if el.Type == notion.BlockTypeHeading2 {
			dto.HTML += fmt.Sprintf("<h2>%v</h2>\n", el.Heading2.Text[0].PlainText)
			continue
		}

		if el.Type == notion.BlockTypeImage {
			imgdto := models.WordPressMediaDTO{
				OriginalUrl: el.Image.File.URL,
			}
			if len(el.Image.Caption) > 0 {
				imgdto.Name = el.Image.Caption[0].PlainText
				imgdto.Slug = slug.Make(imgdto.Name)
			}
			imgdto.Tag = fmt.Sprintf("{{%v}}", imgdto.Slug)
			dto.HTML += imgdto.Tag + "\n"
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

	return dto
}
