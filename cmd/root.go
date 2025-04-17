/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bmoc/cmd/commands/bm"
	"bmoc/cmd/commands/clerk"
	"bmoc/cmd/commands/config"
	"bmoc/cmd/commands/ps"
	"fmt"
	"log"
	"os"
	"os/user"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "bmoc",
	Short: "Personal automation CLI by Brian Morrison II (@brianmmdev)",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Import config
	initConfig()
	rootCmd.AddCommand(config.ConfigCmd)
	rootCmd.AddCommand(ps.PsCmd)
	rootCmd.AddCommand(bm.BmCommand)
	rootCmd.AddCommand(IdCommand)
	rootCmd.AddCommand(clerk.ClerkCmd)
}

func initConfig() {
	myself, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	configFile = fmt.Sprintf("%v/.bmoc.yml", myself.HomeDir)
	viper.SetConfigType("yaml")
	viper.SetConfigFile(configFile)

	viper.AutomaticEnv()

	viper.ReadInConfig()
	// if err != nil {
	// 	log.Fatal("Unable to load configuration file")
	// }
}
