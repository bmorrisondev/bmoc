package container

import (
	"bmoc/cmd/services"
	"fmt"
	"log"

	"github.com/dstotijn/go-notion"
	"github.com/spf13/viper"
)

var (
	notionClient *notion.Client
	wpclient     *services.WordPressClient
)

func GetNotionClient() *notion.Client {
	if notionClient == nil {
		notionKey := viper.GetString("NOTION_KEY")
		if notionKey == "" {
			log.Fatal("ERROR: Config 'NOTION_KEY' is required to be set.")
		}
		notionClient = notion.NewClient(notionKey)
	}
	return notionClient
}

func GetWordPressClient() *services.WordPressClient {
	if wpclient == nil {
		url := viper.GetString("WP_URL")
		username := viper.GetString("WP_USERNAME")
		password := viper.GetString("WP_PASSWORD")

		wpclient = &services.WordPressClient{
			BaseUrl:  fmt.Sprintf("%v/wp-json/wp/v2", url),
			Username: username,
			Password: password,
		}
	}
	return wpclient
}
