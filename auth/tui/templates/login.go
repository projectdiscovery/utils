package templates

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	PlatformName  = "Project Discovery"
	SelectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFFF")) // cyan
	NormalStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff")) // white
	GreyStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#808080")) // grey
	helperTxt     = "Use arrow keys"
	loginFmt      = fmt.Sprintf("%s %s %s %s\n", GreyStyle.Render("?"), NormalStyle.Render("Log in to"), NormalStyle.Render(PlatformName), GreyStyle.Render("(%s)"))
)

// LoginChoice represents a login option with a title and action to execute.
type LoginChoice struct {
	Title  string
	Style  lipgloss.Style
	Action func() tea.Cmd
}

func (c LoginChoice) Init() tea.Cmd {
	return nil
}

func (c LoginChoice) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return c, nil
}

func (c LoginChoice) View() string {
	return c.Title
}

// LoginModel represents the state of the login template.
type LoginModel struct {
	toolName    string
	version     string
	description string
	choices     []LoginChoice
	cursor      int
	selected    bool
}

// NewLoginModel creates a new login model with the provided tool name, version, and choices.
func NewLoginModel(toolName, version, description string, choices []LoginChoice) LoginModel {
	choices = append(choices, LoginChoice{
		Title: "Skip Login",
		Action: func() tea.Cmd {
			return func() tea.Msg {
				return tea.Quit()
			}
		},
	})

	return LoginModel{
		toolName:    toolName,
		version:     version,
		description: description,
		choices:     choices,
	}
}

// Init initializes the login model.
func (m LoginModel) Init() tea.Cmd {
	return nil
}

// Update handles key messages and user input.
func (m LoginModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			if !m.selected && m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if !m.selected && m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter":
			if !m.selected {
				m.selected = true
				return m, m.choices[m.cursor].Action()
			}
		}
	}
	return m, nil
}

// View renders the login template UI.
func (m LoginModel) View() string {
	var preview string
	preview = fmt.Sprintf("%s [Version: %s]\n", NormalStyle.Render(m.toolName), NormalStyle.Render(m.version))

	if m.description != "" {
		preview += fmt.Sprintf("%s %s\n", GreyStyle.Render(">"), NormalStyle.Render(m.description))
	}

	if m.selected {
		return preview + fmt.Sprintf(loginFmt, GreyStyle.Render(m.choices[m.cursor].Title))
	}

	var template string
	template += preview + fmt.Sprintf(loginFmt, GreyStyle.Render(helperTxt))

	for i, choice := range m.choices {
		cursor := " "
		// prepend a divider before skip login
		if i == len(m.choices)-1 {
			template += "─────────────────────────────────\n"
		}
		if m.cursor == i {
			cursor = "❯"
			template += fmt.Sprintf("%s\n", SelectedStyle.Render(cursor, choice.Title))
		} else {
			template += fmt.Sprintf("%s\n", choice.Style.Render(cursor, choice.Title))
		}
	}

	return fmt.Sprintf("\n%s", template)
}

func (m LoginModel) IsLoginSkiped() bool {
	return m.cursor == len(m.choices)-1
}
