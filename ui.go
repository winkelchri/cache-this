package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/lipgloss"
	"github.com/sirupsen/logrus"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

var (
	highlight      = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	highlightStyle = lipgloss.NewStyle().Foreground(highlight).Render

	special      = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}
	specialStyle = lipgloss.NewStyle().Foreground(special).Render

	spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
)

type model struct {
	cacheDir *CacheDir
	keymap   keymap
	typing   bool
	loading  bool
	confirm  bool
	reading  bool
	err      error
	// styles   common.Styles

	help      help.Model
	textInput textinput.Model
	spinner   spinner.Model
	progress  progress.Model
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
	var cmd tea.Cmd

	// Handle key events
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {

		// quit
		case key.Matches(msg, m.keymap.quit):
			return m, tea.Quit

		// enter
		case key.Matches(msg, m.keymap.enter):

			// enter while typing state
			if m.typing {
				m.typing = false

				if path := strings.TrimSpace(m.textInput.Value()); path != "" {
					m.loading = true

					return m, tea.Batch(
						spinner.Tick,
						m.fetchDirectoryInfo(path),
					)
				}
			}

			// enter while reading state
			if m.confirm {
				m.confirm = false
				m.reading = true

				return m, tea.Batch(
					spinner.Tick,
					m.readDirectory(),
				)
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
		m.confirm = true

		if err := msg.Err; err != nil {
			m.err = err
			return m, nil
		}

		m.cacheDir = msg.DirectoryInfo

	case ReadingFinished:
		m.reading = false
		if err := msg.Err; err != nil {
			m.err = err
			return m, nil
		}
	}

	if m.typing {
		m.textInput, cmd = m.textInput.Update(msg)
	}

	if m.loading {
		m.spinner, cmd = m.spinner.Update(msg)
	}

	if m.reading {
		m.spinner, cmd = m.spinner.Update(msg)
	}

	return m, cmd
}

// View is responsible for rendering screen information.
func (m model) View() string {
	log.SetLevel(logrus.ErrorLevel)
	var s string

	// typing mode
	if m.typing {
		s += "Enter directory:\n"
		s += m.textInput.View()
	}

	// loading mode
	if m.loading {
		s += m.spinner.View()
		s += " loading directory info. Please wait..."
	}

	// error mode
	if err := m.err; err != nil {
		s += "Error fetching directory info: "
		s += err.Error()
	}

	// reading directory contents
	if m.reading {
		// n := m.cacheDir.numFiles
		// w := lipgloss.Width(fmt.Sprintf("%d", n))
		s = m.progress.View() + "\n\n"
		s += m.spinner.View() + "Reading files. Please wait..."
	}

	// default state
	if s == "" {
		s = fmt.Sprintf(
			"Found %s files in '%s'.\nTotal size: %.2f MB.",
			highlightStyle(strconv.FormatInt(m.cacheDir.numFiles, 10)),
			specialStyle(m.cacheDir.path),
			SizeInMB(m.cacheDir.sizeDir),
		)
	}

	s += "\n" + m.helpView()

	return s
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

/*
TODO: 	Animate readDirectory. Spinner is boring!
*/

type ReadingFinished struct {
	Err  error
	Done bool
}

func (m model) readDirectory() tea.Cmd {
	// TODO: 	Possibly rewrite with go routines.
	// 			See bubbletea/examples/send-msg

	return func() tea.Msg {
		var err error
		for _, each := range m.cacheDir.files {
			err = each.Read()
		}

		if err != nil {
			return ReadingFinished{Err: err, Done: false}
		}

		return ReadingFinished{Done: true}
	}
}

func initialModel() model {
	t := textinput.NewModel()
	t.Focus()
	t.SetValue(os.Getenv("CACHE_THIS_DIR"))

	s := spinner.NewModel()
	s.Spinner = spinner.Dot
	s.Style = spinnerStyle

	pr := progress.New(progress.WithDefaultGradient())

	m := model{
		textInput: t,
		spinner:   s,
		typing:    true,
		progress:  pr,
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
