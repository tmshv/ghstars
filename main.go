package main

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type (
	errMsg error
)

type model struct {
	textInput textinput.Model
	list      list.Model
	err       error
}

func OpenURL(url string) error {
	cmd := exec.Command("open", url) // Use "xdg-open" on Linux or "start" on Windows
	err := cmd.Start()
	if err != nil {
		return err
	}

	// Wait for the command to finish executing
	err = cmd.Wait()
	if err != nil {
		return err
	}

	return nil
}

func initialModel() model {
	var style = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4"))
		// Bold(true).
		// PaddingTop(2).
		// PaddingLeft(4).
		// Width(22)

	ti := textinput.New()
	ti.Placeholder = "Query"
	ti.TextStyle = style
	ti.PlaceholderStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#BABABA")).
		Background(lipgloss.Color("#7D56F4"))
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20

	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Stars"

	return model{
		textInput: ti,
		list:      l,
		err:       nil,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			c := m.list.Cursor()
			i := m.list.Items()[c]
			switch val := i.(type) {
			case repoitem:
				err := OpenURL(val.URL())
				if err != nil {
					return m, tea.Quit
				}
			}
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)

	case AddStarMsg:
		i := repoitem{
			url:   msg.star.Repo.HTMLURL,
			title: msg.star.Repo.HTMLURL,
			desc:  msg.star.Repo.Description,
		}
		m.list.InsertItem(0, i)
		return m, nil

	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	m.textInput, cmd = m.textInput.Update(msg)
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	repos := docStyle.Render(m.list.View())
	search := m.textInput.Value()
	return fmt.Sprintf(
		"Query: \n\n%s\n\nIs=%s\n\n%s\n\n%s",
		m.textInput.View(),
		search,
		repos,
		"(esc to quit)",
	) + "\n"
}

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type repoitem struct {
	url, title, desc string
}

func (i repoitem) URL() string         { return i.url }
func (i repoitem) Title() string       { return i.title }
func (i repoitem) Description() string { return i.desc }
func (i repoitem) FilterValue() string { return i.title }

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())

	go Ghfetch(p)

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
