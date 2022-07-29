package utils

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type Choice struct {
	Id   string
	Name string
}

type Options struct {
	Choices  []Choice
	Cursor   int
	Callback func(key string)
}

func (m Options) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (o Options) View() string {
	s := "What article should be pushed to WordPress?\n\n"

	for i, choice := range o.Choices {
		cursor := " "
		if o.Cursor == i {
			cursor = ">"
		}
		s += fmt.Sprintf("%s %s\n", cursor, choice.Name)
	}

	s += "\nPress q to quit.\n"
	return s
}

func (o Options) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {

		case "ctrl+c", "q":
			return o, tea.Quit

		case "up", "k":
			if o.Cursor > 0 {
				o.Cursor--
			}

		case "down", "j":
			if o.Cursor < len(o.Choices)-1 {
				o.Cursor++
			}

		case "enter":
			selected := o.Choices[o.Cursor]
			o.Callback(selected.Id)
			return o, tea.Quit
		}
	}

	return o, nil
}

func PresentSelector(options Options) {
	p := tea.NewProgram(&options)
	p.Start()
}
