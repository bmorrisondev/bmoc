/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package config

import (
	"bmoc/cmd/utils"
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

var (
	delKey string
)

// deleteCmd represents the delete command
var DeleteCmd = &cobra.Command{
	Use: "delete",
	Run: func(cmd *cobra.Command, args []string) {
		if delKey == "" {
			log.Fatal("'key' must be specified")
		}
		fmt.Printf("\n\n    **** Deleting key: %s ****\n\n", delKey)
		utils.ConfigKeyValuePairDelete(delKey)
	},
}

func init() {
	DeleteCmd.Flags().StringVarP(&delKey, "key", "k", "", "Config key")
}
