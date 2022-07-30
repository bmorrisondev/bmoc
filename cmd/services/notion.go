package services

import (
	"bmoc/cmd/models"
	"bmoc/cmd/utils"
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/dstotijn/go-notion"
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

func SetupDoc(contentItemId string) {
	setup()

	// Get it again
	page, err := client.FindPageByID(context.TODO(), contentItemId)
	if err != nil {
		log.Fatal(err)
	}
	props := page.Properties.(notion.DatabasePageProperties)
	pageName := props["Name"].Title[0].PlainText
	areaId := props["Area"].Relation[0].ID
	types := []string{}
	for _, el := range props["Type"].MultiSelect {
		types = append(types, el.Name)
	}

	area, err := client.FindPageByID(context.TODO(), areaId)
	if err != nil {
		log.Fatal(err)
	}
	areaProps := area.Properties.(notion.DatabasePageProperties)
	areaName := areaProps["Name"].Title[0].PlainText

	projectTasks := models.BuildProjectTaskList(areaName, types)

	// Update page with icon & draft status
	updParams := notion.UpdatePageParams{
		// TODO: Grab the icon from the Area and apply it here
		Icon: &notion.Icon{
			File: area.Icon.File,
		},
		DatabasePageProperties: notion.DatabasePageProperties{
			"Status": notion.DatabasePageProperty{
				Select: &notion.SelectOptions{
					Name: "Ready",
				},
			},
		},
	}
	_, err = client.UpdatePage(context.TODO(), contentItemId, updParams)
	if err != nil {
		log.Println("Updating page")
		log.Fatal(err)
	}

	// Create project
	dbid := utils.GetConfigString("NOTION_PROJECTS_DB", true)
	params := notion.CreatePageParams{
		ParentType: notion.ParentTypeDatabase,
		ParentID:   dbid,
		DatabasePageProperties: &notion.DatabasePageProperties{
			"Name": notion.DatabasePageProperty{
				Title: []notion.RichText{
					{
						Text: &notion.Text{
							Content: pageName,
						},
					},
				},
			},
			"Status": notion.DatabasePageProperty{
				Select: &notion.SelectOptions{
					Name: "Active",
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

	projectPage, err := client.CreatePage(context.TODO(), params)
	if err != nil {
		log.Panic(err)
	}
	log.Println("Created project:", projectPage.ID)

	// Create tasks
	dbid = utils.GetConfigString("NOTION_TASKS_DB", true)
	for task, subtasks := range projectTasks {
		params = notion.CreatePageParams{
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
				"Project": notion.DatabasePageProperty{
					Relation: []notion.Relation{
						{
							ID: projectPage.ID,
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
	}
}
