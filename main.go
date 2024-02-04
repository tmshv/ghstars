package main

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

type (
	errMsg error
)

type repoitem struct {
	url      string
	title    string
	desc     string
	tags     []string
	lang     string
	archived bool
}

func (i repoitem) URL() string         { return i.url }
func (i repoitem) Title() string       { return i.title }
func (i repoitem) Description() string { return i.desc }
func (i repoitem) FilterValue() string {
	var blocks []string
	blocks = append(blocks, i.url)
	blocks = append(blocks, i.desc)
	blocks = append(blocks, i.lang)
	blocks = append(blocks, i.tags...)
	return strings.Join(blocks, " ")
}

type model struct {
	username     string
	items        []repoitem
	showArchived bool
	textInput    textinput.Model
	list         list.Model
	keys         *listKeyMap
	err          error
}

func (m *model) getItems() []list.Item {
	items := make([]list.Item, 0, len(m.items))
	for _, item := range m.items {
		if m.showArchived || !item.archived {
			items = append(items, item)
		}
	}
	return items
}

var (

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1)

	docStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63"))
)

var (
	username string
	useCache bool
)

type listKeyMap struct {
	toggleShowArchived key.Binding
}

func newListKeyMap() *listKeyMap {
	return &listKeyMap{
		toggleShowArchived: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "show archived"),
		),
	}
}

func monthsPassed(t time.Time) int {
	// f := t.Format("20060102")
	return int(time.Since(t).Hours() / 24 / 30)
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

func initialModel(username string) model {
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

	var (
		listKeys = newListKeyMap()
	)

	listdelegate := list.NewDefaultDelegate()
	listdelegate.ShowDescription = true
	listdelegate.SetHeight(3)
	l := list.New([]list.Item{}, listdelegate, 0, 0)
	// l.SetFilteringEnabled(false)
	l.Title = fmt.Sprintf("%s's Stars", username)
	l.Styles.Title = titleStyle
	l.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			listKeys.toggleShowArchived,
		}
	}

	return model{
		username:  username,
		textInput: ti,
		list:      l,
		keys:      listKeys,
		err:       nil,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		GhStartFetch,
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Don't match any of the keys below if we're actively filtering.
		if m.list.FilterState() == list.Filtering {
			break
		}

		switch {
		case key.Matches(msg, m.keys.toggleShowArchived):
			m.showArchived = !m.showArchived
			cmd := m.list.SetItems(m.getItems())
			return m, cmd
		}

		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyEnter:
			val, ok := m.list.SelectedItem().(repoitem)
			if ok {
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
		// Skip archived repo (be an option in the future)

		var upd string
		last := monthsPassed(msg.star.Repo.UpdatedAt)
		if last == 0 {
			upd = "active"
		} else {
			upd = fmt.Sprintf("last %dm", last)
		}
		if msg.star.Repo.Archived {
			upd = "archived"
		}
		desc := fmt.Sprintf("(%s; added %d; stars %d) \n %s",
			upd,
			monthsPassed(msg.star.StarredAt),
			msg.star.Repo.StargazersCount,
			msg.star.Repo.Description,
		)
		i := repoitem{
			url:      msg.star.Repo.HTMLURL,
			title:    msg.star.Repo.HTMLURL,
			desc:     desc,
			lang:     msg.star.Repo.Language,
			tags:     msg.star.Repo.Topics,
			archived: msg.star.Repo.Archived,
		}
		m.items = append(m.items, i)
		m.list.InsertItem(10000, i) // TODO use other value to add at the end of list
		return m, nil

	case GhstarsStartMsg:
		cmd := m.list.StartSpinner()
		return m, cmd

	case GhstarsStopMsg:
		m.list.StopSpinner()

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
	var blocks []string

	// value := m.textInput.Value()

	// blocks = append(blocks, docStyle.Render(m.list.View()))
	// blocks = append(blocks, m.textInput.View())
	blocks = append(blocks, m.list.View())
	val, ok := m.list.SelectedItem().(repoitem)
	if ok {
		blocks = append(blocks, docStyle.Render(fmt.Sprintf("select: %s", val.url)))
	}

	return m.list.View()

	return lipgloss.JoinVertical(lipgloss.Left, blocks...)
}

func parseCLI() *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:   "ghstars",
		Short: "ghstars fetches and displays GitHub stars for a user",
		Long:  `ghstars is a CLI application that fetches and displays the GitHub stars for a specified user`,
		RunE: func(cmd *cobra.Command, args []string) error {
			m := initialModel(username)
			p := tea.NewProgram(m, tea.WithAltScreen())

			go Ghfetch(p, username, useCache)

			_, err := p.Run()
			return err
		},
	}

	rootCmd.Flags().StringVarP(&username, "username", "u", "", "GitHub username to fetch stars for")
	rootCmd.MarkFlagRequired("username")

	rootCmd.Flags().BoolVarP(&useCache, "cache", "c", false, "Use cached data instead of fetching new data")

	return rootCmd
}

func main() {
	rootCmd := parseCLI()
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
