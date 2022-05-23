package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	cacheDir *CacheDir
	keymap   keymap
	typing   bool
	loading  bool
	err      error
	// styles   common.Styles

	help      help.Model
	textInput textinput.Model
	spinner   spinner.Model
}

type keymap struct {
	enter key.Binding
	back  key.Binding
	quit  key.Binding
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) helpView() string {
	return "\n" + m.help.ShortHelpView([]key.Binding{
		m.keymap.enter,
		m.keymap.back,
		m.keymap.quit,
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	var (
		cmd tea.Cmd
	)

	// Handle key events
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {

		// quit
		case key.Matches(msg, m.keymap.quit):
			return m, tea.Quit

		// enter
		case key.Matches(msg, m.keymap.enter):
			if m.typing {
				if path := strings.TrimSpace(m.textInput.Value()); path != "" {
					m.typing = false
					m.loading = true

					return m, tea.Batch(
						spinner.Tick,
						m.fetchDirectoryInfo(path),
					)
				}
			}

		// back
		case key.Matches(msg, m.keymap.back):
			if !m.typing && !m.loading {
				m.typing = true
				m.err = nil
			}
		}

	// directory information returned
	case GotDirectoryInfo:
		m.loading = false

		if err := msg.Err; err != nil {
			m.err = err
			return m, nil
		}

		m.cacheDir = msg.DirectoryInfo
	}

	if m.typing {
		m.textInput, cmd = m.textInput.Update(msg)
	}

	if m.loading {
		m.spinner, cmd = m.spinner.Update(msg)
	}

	return m, cmd
}

func (m model) View() string {

	// typing mode
	if m.typing {
		return fmt.Sprintf(
			"Enter directory:\n%s\n%s",
			m.textInput.View(),
			m.helpView(),
		)
	}

	// loading mode
	if m.loading {
		return fmt.Sprintf(
			"%s loading directory info. Please wait...",
			m.spinner.View(),
		)
	}

	// error mode
	if err := m.err; err != nil {
		return fmt.Sprintf(
			"Error fetching directory info: %v\n%s",
			err,
			m.helpView(),
		)
	}

	// default state
	return fmt.Sprintf(
		"Found %d files in '%s'. Total size: %.2f MB.\n%s",
		m.cacheDir.numFiles,
		m.cacheDir.path,
		SizeInMB(m.cacheDir.sizeDir),
		m.helpView(),
	)
}

type GotDirectoryInfo struct {
	Err           error
	DirectoryInfo *CacheDir
}

func (m model) fetchDirectoryInfo(path string) tea.Cmd {
	return func() tea.Msg {
		c, err := GetDirectoryInfo(path)
		if err != nil {
			return GotDirectoryInfo{Err: err}
		}

		return GotDirectoryInfo{DirectoryInfo: &c}
	}
}

func initialModel() model {
	t := textinput.NewModel()
	t.Focus()
	t.SetValue(os.Getenv("CACHE_THIS_DIR"))

	s := spinner.NewModel()
	s.Spinner = spinner.Dot

	m := model{
		textInput: t,
		spinner:   s,
		typing:    true,
		keymap: keymap{
			enter: key.NewBinding(
				key.WithKeys("enter"),
				key.WithHelp("enter", "confirm"),
			),
			back: key.NewBinding(
				key.WithKeys("esc"),
				key.WithHelp("esc", "back"),
			),
			quit: key.NewBinding(
				key.WithKeys("ctrl+c", "q"),
				key.WithHelp("q", "quit"),
			),
		},
		help: help.NewModel(),
	}

	// m.keymap.back.SetEnabled(false)

	return m
}

func StartUI() {
	// if err := tea.NewProgram(initialModel(), tea.WithAltScreen()).Start(); err != nil {
	if err := tea.NewProgram(initialModel()).Start(); err != nil {
		fmt.Printf("could not start program: %s\n", err)
		os.Exit(1)
	}
}
