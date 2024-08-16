package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/timer"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

var duration = []string{"30 Min ⌛", "45 Min ⌛", "1 Hr ⌛"}

type (
	errMsg error
)

type model struct {
	timer    timer.Model
	keymap   keymap
	help     help.Model
	quitting bool
	textInput textinput.Model
	cursor   int
	choice   string
	err      error
	timeout  time.Duration // Store the selected timeout duration
}

type keymap struct {
	start key.Binding
	stop  key.Binding
	reset key.Binding
	quit  key.Binding
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Pikachu"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20

	return model{
		keymap: keymap{
			start: key.NewBinding(
				key.WithKeys("s"),
				key.WithHelp("s", "start"),
			),
			stop: key.NewBinding(
				key.WithKeys("s"),
				key.WithHelp("s", "stop"),
			),
			reset: key.NewBinding(
				key.WithKeys("r"),
				key.WithHelp("r", "reset"),
			),
			quit: key.NewBinding(
				key.WithKeys("q", "ctrl+c"),
				key.WithHelp("q", "quit"),
			),
		},
		help:     help.New(),
		textInput: ti,
		err:      nil,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) helpView() string {
	return "\n" + m.help.ShortHelpView([]key.Binding{
		m.keymap.start,
		m.keymap.stop,
		m.keymap.reset,
		m.keymap.quit,
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			m.choice = duration[m.cursor]

			// Set the timer duration based on the selected option
			switch m.choice {
			case "30 Min ⌛":
				m.timeout = time.Second * 1800
			case "45 Min ⌛":
				m.timeout = time.Second * 2700
			case "1 Hr ⌛":
				m.timeout = time.Second * 3600
			}

			// Initialize the timer with the selected duration
			m.timer = timer.NewWithInterval(m.timeout, time.Millisecond)

			// Start the timer
			return m, m.timer.Init()

		case "down", "j":
			m.cursor++
			if m.cursor >= len(duration) {
				m.cursor = 0
			}
		case "up", "k":
			m.cursor--
			if m.cursor < 0 {
				m.cursor = len(duration) - 1
			}
		}

		// Handle other key matches outside of the switch statement
		if key.Matches(msg, m.keymap.quit) {
			m.quitting = true
			return m, tea.Quit
		}

		if key.Matches(msg, m.keymap.reset) {
			// Reset the timer by reinitializing it
			m.timer = timer.NewWithInterval(m.timeout, time.Millisecond)
			return m, m.timer.Init()
		}

		if key.Matches(msg, m.keymap.start, m.keymap.stop) {
			return m, m.timer.Toggle()
		}

	// Handle timer messages
	case timer.TickMsg:
		var cmd tea.Cmd
		m.timer, cmd = m.timer.Update(msg)
		return m, cmd

	case timer.StartStopMsg:
		var cmd tea.Cmd
		m.timer, cmd = m.timer.Update(msg)
		m.keymap.stop.SetEnabled(m.timer.Running())
		m.keymap.start.SetEnabled(!m.timer.Running())
		return m, cmd

	case timer.TimeoutMsg:
		m.quitting = true
		return m, tea.Quit
	}

	return m, nil
}

func (m model) View() string {
	s := strings.Builder{}
	s.WriteString("\t\t\t\t 🍅 Pomodoro 🍅\n \n")
	s.WriteString("\t\t\t Select Timer⏰ :\n \n")

	for i := 0; i < len(duration); i++ {
		if m.cursor == i {
			s.WriteString("\t\t\t->> ")
		} else {
			s.WriteString("\t\t\t( ) ")
		}
		s.WriteString("\t"+duration[i])
		s.WriteString("\n")
	}

	s.WriteString("\n(Press Q to quit)\n\n")

	// Timer view
	timerView := m.timer.View()

	if m.timer.Timedout() {
		timerView = "\t\t\t All done!"
	}

	// Combine the selection view with the timer view
	s.WriteString("\t\t\t"+timerView + "\n")

	if !m.quitting {
		s.WriteString("Exiting in " + timerView + "\n")
		s.WriteString(m.helpView())
	}

	return s.String()
}

func main() {
	p := tea.NewProgram(initialModel())

	// Run returns the model as a tea.Model.
	m, err := p.Run()
	if err != nil {
		fmt.Println("Oh no:", err)
		os.Exit(1)
	}

	// Assert the final tea.Model to our local model and print the choice.
	if m, ok := m.(model); ok && m.choice != "" {
		fmt.Printf("\nTimer started for %s\n", m.choice)
	}
}
