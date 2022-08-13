package cmd

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/teris-io/shortid"
)

var IdCommand = &cobra.Command{
	Use:   "id",
	Short: "Generates an id",
	Run: func(cmd *cobra.Command, args []string) {
		sid, err := shortid.New(1, shortid.DefaultABC, 2342)
		if err != nil {
			log.Fatal(err)
		}
		id, err := sid.Generate()
		if err != nil {
			log.Fatal(err)
		}
		lowered := strings.ToLower(id)
		year, _, _ := time.Now().Date()
		fmt.Printf("%v_%v\n", year, lowered)
	},
}
