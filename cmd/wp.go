/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"bmoc/cmd/container"
	"bmoc/cmd/services"
	"context"
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dstotijn/go-notion"
	"github.com/gosimple/slug"
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
			MoveArticleToWordPress(selected.Id)
			return m, tea.Quit
		}
	}

	return m, nil
}

type WordPressPageDTO struct {
	HTML           string
	Excerpt        string
	ImagesToUpload []WordPressMediaDTO
}

type WordPressMediaDTO struct {
	Name        string
	OriginalUrl string
	Alt         string
}

func MoveArticleToWordPress(pageId string) {
	dto := WordPressPageDTO{
		ImagesToUpload: []WordPressMediaDTO{},
	}

	// Get the page from Notion
	client := container.GetNotionClient()
	// TODO: Need to grab the excerpt from here
	// page, err := client.FindPageByID(context.TODO(), pageId)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	blocks, err := client.FindBlockChildrenByID(context.TODO(), pageId, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Iterate over each block and create html
	for _, el := range blocks.Results {
		if el.Type == notion.BlockTypeHeading1 {
			dto.HTML += fmt.Sprintf("<h1>%v</h1>\n", el.Heading1.Text[0].PlainText)
		}

		if el.Type == notion.BlockTypeHeading2 {
			dto.HTML += fmt.Sprintf("<h2>%v</h2>\n", el.Heading2.Text[0].PlainText)
		}

		if el.Type == notion.BlockTypeImage {
			imgdto := WordPressMediaDTO{
				OriginalUrl: el.Image.File.URL,
			}
			if len(el.Image.Caption) > 0 {
				imgdto.Name = el.Image.Caption[0].PlainText
				imgdto.Alt = slug.Make(imgdto.Name)
			}
			// TODO: swap out original URL, implement captions
			dto.HTML += fmt.Sprintf("<img src=\"%v\" alt=\"%v\" />\n", imgdto.OriginalUrl, imgdto.Alt)
			dto.ImagesToUpload = append(dto.ImagesToUpload, imgdto)
			continue
		}

		if el.Type == notion.BlockTypeParagraph && len(el.Paragraph.Text) > 0 {
			dto.HTML += "<p>"
			for _, ptext := range el.Paragraph.Text {
				if ptext.Annotations.Bold {
					dto.HTML += fmt.Sprintf("<b>%v</b>", ptext.PlainText)
					continue
				}

				if ptext.Annotations.Italic {
					dto.HTML += fmt.Sprintf("<i>%v</i>", ptext.PlainText)
					continue
				}

				if ptext.Annotations.Code {
					// TODO: figure this one out
					continue
				}

				dto.HTML += ptext.PlainText
			}
			dto.HTML += "</p>\n"
		}

	}

	wpclient := container.GetWordPressClient()
	uplurl, err := wpclient.UploadMediaFromUrl(dto.ImagesToUpload[0].OriginalUrl, dto.ImagesToUpload[0].Name, dto.ImagesToUpload[0].Name)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(*uplurl)
}
