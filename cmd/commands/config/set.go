/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package config

import (
	"bmoc/cmd/utils"

	"github.com/spf13/cobra"
)

var (
	setKey string
	setVal string
)

// addCmd represents the add command
var SetCmd = &cobra.Command{
	Use: "set",
	Run: func(cmd *cobra.Command, args []string) {
		utils.ConfigKeyValuePairAdd(setKey, setVal)
	},
}

func init() {
	SetCmd.Flags().StringVarP(&setKey, "key", "k", "", "Config key")
	SetCmd.Flags().StringVarP(&setVal, "value", "v", "", "Config value")
}
