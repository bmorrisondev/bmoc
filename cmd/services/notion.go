package services

import (
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
