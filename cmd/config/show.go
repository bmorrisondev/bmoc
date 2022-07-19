/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package config

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// addCmd represents the add command
var ShowCmd = &cobra.Command{
	Use: "show",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("** All keys including environment variables for CLI.\n")
		fmt.Printf("%s\n\n", viper.AllKeys())

		settings := viper.AllSettings()
		fmt.Printf("** Configuration file keys and values.\n")
		for i, v := range settings {
			fmt.Printf("%v: %v\n", i, v)
		}
	},
}
