package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	lib "github.com/charmbracelet/charm/ui/common"
	te "github.com/muesli/termenv"
)

var (
	color                        = te.ColorProfile().Color
	textFocusedColor             = lib.Color("#EE6FF8")
	buttonBackgroundColor        = lib.Color("#6E2D73")
	buttonBackgroundFocusedColor = lib.Color("#EE6FF8")
	buttonForegroundColor        = lib.Color("#FFFFFF")
	// focusedPrompt       = te.String("• ").Foreground(color("205")).String()
	// blurredPrompt       = "• "
	// focusedSubmitButton = "[ " + te.String("Add").Foreground(color("205")).String() + " ]"
	// blurredSubmitButton = "[ " + te.String("Add").Foreground(color("240")).String() + " ]"
)

func textBlurred(text string) string {
	return te.String(text).Foreground(color("255")).String()
}

func textFocused(text string) string {
	return te.String(text).Foreground(textFocusedColor).String()
}

func buttonBlurred(text string) string {
	return te.String(" " + text + " ").
		Bold().
		Background(buttonBackgroundColor).
		Foreground(buttonForegroundColor).
		String()

	// return "[ " + te.String(text).Foreground(color("255")).String() + " ]"
}

func buttonFocused(text string) string {
	return te.String(" " + text + " ").
		Bold().
		Background(buttonBackgroundFocusedColor).
		Foreground(buttonForegroundColor).
		String()

	// return "[ " + te.String(text).Foreground(color("205")).String() + " ]"
}

type model struct {
	aliasInput        textinput.Model
	publicKeyInput    textinput.Model
	relationships     map[string]string //*relationship.Relationship
	relationshipsList []string
	cursor            int
	selected          map[int]struct{}
}

func initialModel() model {
	aliasInput := textinput.NewModel()
	aliasInput.Focus()
	aliasInput.Placeholder = "a unique alias for this contact"
	aliasInput.Prompt = ""
	publicKeyInput := textinput.NewModel()
	publicKeyInput.Placeholder = "contact's public key"
	publicKeyInput.Prompt = ""
	return model{
		aliasInput:        aliasInput,
		publicKeyInput:    publicKeyInput,
		relationships:     map[string]string{}, //*relationship.Relationship{},
		relationshipsList: []string{},
		cursor:            -3,
		selected:          map[int]struct{}{},
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "up":
			if m.cursor > -3 {
				m.cursor--
			}
			if m.cursor != -3 {
				m.aliasInput.Blur()
			} else {
				m.aliasInput.Focus()
			}
			if m.cursor != -2 {
				m.publicKeyInput.Blur()
			} else {
				m.publicKeyInput.Focus()
			}
			if len(m.relationships) == 0 {
				break
			}
		case "down":
			if m.cursor+1 < len(m.relationships) {
				m.cursor++
			}
			if m.cursor != -3 {
				m.aliasInput.Blur()
			} else {
				m.aliasInput.Focus()
			}
			if m.cursor != -2 {
				m.publicKeyInput.Blur()
			} else {
				m.publicKeyInput.Focus()
			}
			if len(m.relationships) == 0 {
				break
			}
		case "enter":
			if m.cursor == -3 {
				m.cursor++
				m.aliasInput.Blur()
				m.publicKeyInput.Focus()
				break
			}
			if m.cursor == -2 {
				m.publicKeyInput.Blur()
				m.cursor++
				break
			}
			if m.cursor == -1 {
				aliasInput, _ := m.aliasInput.Update(msg)
				publicKeyInput, _ := m.publicKeyInput.Update(msg)
				alias := strings.TrimSpace(aliasInput.Value())
				publicKey := strings.TrimSpace(publicKeyInput.Value())
				if alias == "" || publicKey == "" {
					break
				}
				m.relationships[alias] = publicKey
				m.aliasInput.Reset()
				m.aliasInput.Blur()
				m.publicKeyInput.Reset()
				m.publicKeyInput.Blur()
				exists := false
				for _, v := range m.relationshipsList {
					if v == alias {
						exists = true
						break
					}
				}
				if !exists {
					m.relationshipsList = append(m.relationshipsList, alias)
				}
				return m, nil
			}
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}
		}
	}

	var cmd tea.Cmd
	if m.cursor == -3 {
		m.aliasInput, cmd = m.aliasInput.Update(msg)
		return m, cmd
	}
	if m.cursor == -2 {
		m.publicKeyInput, cmd = m.publicKeyInput.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m model) View() string {
	s := "Add contact:\n"

	if m.cursor == -3 {
		s += "• " + textFocused("Alias: ")
	} else {
		s += "  " + textBlurred("Alias: ")
	}

	s += m.aliasInput.View() + "\n"

	if m.cursor == -2 {
		s += "• " + textFocused("Public key: ")
	} else {
		s += "  " + textBlurred("Public key: ")
	}

	s += m.publicKeyInput.View() + "\n"

	if m.cursor == -1 {
		s += "• " + buttonFocused("add") + "\n\n"
	} else {
		s += "  " + buttonBlurred("add") + "\n\n"
	}

	if len(m.relationshipsList) == 0 {
		s += "No contacts.\n"
	} else {
		s += "Contacts:\n"
		for i, alias := range m.relationshipsList {
			publicKey := m.relationships[alias]
			if m.cursor == i {
				s += fmt.Sprintf("• %s [%s]\n", textFocused("@"+alias), publicKey)
			} else {
				s += fmt.Sprintf("  %s [%s]\n", textBlurred("@"+alias), publicKey)
			}
		}
	}

	s += "\nPress ctrl+c to quit.\n"

	return s
}

func main() {
	p := tea.NewProgram(initialModel())
	if err := p.Start(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
