package utils

import (
	"log"

	"github.com/spf13/viper"
)

func GetConfigString(key string, errorOnMissingConfig bool) string {
	val := viper.GetString(key)
	if val == "" && errorOnMissingConfig {
		log.Fatal("ERROR: Config 'NOTION_PROJECTS_DB' is required to be set.")
	}
	return val
}
