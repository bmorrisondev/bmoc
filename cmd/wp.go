/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"bmoc/cmd/services"
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dstotijn/go-notion"
	"github.com/spf13/cobra"
)

// wpCmd represents the wp command
var wpCmd = &cobra.Command{
	Use:   "wp",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		pages, err := services.ListDraftArticles()
		if err != nil {
			log.Fatal(err)
		}

		m := model{
			choices: []Choice{},
		}

		for _, el := range pages {
			c := Choice{Id: el.ID}
			props := el.Properties.(notion.DatabasePageProperties)
			c.Name = props["Name"].Title[0].PlainText
			m.choices = append(m.choices, c)
		}

		p := tea.NewProgram(&m)
		p.Start()
	},
}

func init() {
	rootCmd.AddCommand(wpCmd)
}

type Choice struct {
	Id   string
	Name string
}

type model struct {
	choices []Choice
	cursor  int
}

func (m model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m model) View() string {
	s := "What article should be pushed to WordPress?\n\n"

	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		s += fmt.Sprintf("%s %s\n", cursor, choice.Name)
	}

	s += "\nPress q to quit.\n"
	return s
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {

		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}

		case "enter":
			// TODO: Just call the necessary bits here
			selected := m.choices[m.cursor]
			log.Println("Publishing: ", selected.Id, selected.Name)
			return m, tea.Quit
		}
	}

	return m, nil
}

func MoveArticleToWordPress(pageId string) {

}
