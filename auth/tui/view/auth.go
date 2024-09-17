package view

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	tui "github.com/projectdiscovery/utils/auth/tui/handler"
	"github.com/projectdiscovery/utils/auth/tui/templates"
)

// LoginView represents the login view
type LoginView struct {
	model templates.LoginModel
}

// NewLoginView initializes the login view with toolName and version
func NewLoginView(toolName, version string) LoginView {
	// Define the choices and actions here
	choices := []templates.LoginChoice{
		{
			Title: "Continue with GitHub",
			Style: templates.NormalStyle,
			Action: func() tea.Cmd {
				// Perform OAuth and return API key
				return tea.Batch(

					func() tea.Msg {
						res, err := tui.InitiateGitHubOAuth()
						fmt.Println("err:", err)
						if err != nil {
							return OAuthErrorMsg{Err: err}
						}
						return OAuthSuccessMsg{ApiKey: res["api_key"].(string)}
					},
				)
			},
		},
		{
			Title: "Continue with Email",
			Style: templates.GreyStyle,
			Action: func() tea.Cmd {
				return func() tea.Msg {
					return "Selected Email"
				}
			},
		},
		{
			Title: "Continue with SAML Single Sign-On",
			Style: templates.GreyStyle,
			Action: func() tea.Cmd {
				return func() tea.Msg {
					return "Selected Continue with SAML Single Sign-On"
				}
			},
		},
	}

	descriptiveText := "No existing credentials found. Please log in:"
	model := templates.NewLoginModel(toolName, version, descriptiveText, choices)
	return LoginView{model: model}
}

// OAuthSuccessMsg is a message type to hold the API key
type OAuthSuccessMsg struct {
	ApiKey string
}

// OAuthErrorMsg is a message type to hold an error
type OAuthErrorMsg struct {
	Err error
}

// Init initializes the login view
func (v LoginView) Init() tea.Cmd {
	return v.model.Init()
}

// Update handles key messages and user input
func (v LoginView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case OAuthSuccessMsg:
		return v, tea.Quit
	case OAuthErrorMsg:
		return v, tea.Quit
	}

	m, cmd := v.model.Update(msg)
	v.model = m.(templates.LoginModel)
	return v, cmd
}

func (v LoginView) View() string {
	return v.model.View()
}
