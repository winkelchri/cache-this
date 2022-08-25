package main

/*
	Heavily borrowed from https://github.com/charmbracelet/bubbletea. Such an awesome library! <3
*/

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

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
	program   *tea.Program
	highlight = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	special   = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}

	highlightStyle = lipgloss.NewStyle().Foreground(highlight).Render
	specialStyle   = lipgloss.NewStyle().Foreground(special).Render
	spinnerStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	helpStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Margin(1, 0)
	dotStyle       = helpStyle.Copy().UnsetMargins()
	durationStyle  = dotStyle.Copy()
)

const (
	numDisplayResults = 10
)

type resultMsg struct {
	duration   time.Duration
	file       string
	err        error
	done       bool
	filesTotal int
	filesRead  int
}

func (r resultMsg) String() string {
	var s string

	if r.file == "" {
		s = dotStyle.Render(strings.Repeat(".", 30))
	}

	if r.err != nil {
		s = fmt.Sprintf("⚠️ Error reading '%s': %s", highlightStyle(r.file), r.err.Error())
	}

	if s == "" {
		s = fmt.Sprintf(
			"🔍 (%d/%d) Reading '%s' took %s",
			r.filesRead,
			r.filesTotal,
			highlightStyle(r.file),
			durationStyle.Render(r.duration.String()),
		)
	}

	return s
}

type model struct {
	cacheDir *CacheDir
	keymap   keymap
	typing   bool
	loading  bool
	confirm  bool
	reading  bool
	done     bool
	err      error
	results  []resultMsg
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

// Update function will be run upon various different events
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

				m.readDirectory()
				return m, tea.Batch(
					spinner.Tick,
				)
			}

		// back
		case key.Matches(msg, m.keymap.back):
			if !m.typing && !m.loading {
				m.typing = true
				m.err = nil
			}
		}

	// Not really sure if we need this
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	// FrameMsg is sent when the progress bar wants to animate itself
	case progress.FrameMsg:
		newModel, cmd := m.progress.Update(msg)
		if newModel, ok := newModel.(progress.Model); ok {
			m.progress = newModel
		}
		return m, cmd

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

	case resultMsg:
		if msg.done {
			m.done = true
			return m, tea.Quit
		}

		m.results = append(m.results[1:], msg)

		cmd := m.progress.SetPercent(float64(msg.filesRead)/float64(msg.filesTotal) - 1)
		if msg.filesRead >= msg.filesTotal {
			m.reading = false
			// cmd = func() tea.Msg {
			// 	return ReadingFinished{Done: true}
			// }
		}
		return m, tea.Batch(
			spinner.Tick,
			cmd,
		)
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

	if m.done {
		return fmt.Sprintf("Done reading %d files.\n", len(m.cacheDir.files))
	}

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
		// s = m.progress.View() + "\n\n"
		s += m.spinner.View() + "Reading files. Please wait...\n\n"

		for _, res := range m.results {
			s += res.String() + "\n"
		}
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

// func (m model) readDirectory() tea.Cmd {
func (m model) readDirectory() {
	// FIXME: 	Subfunction is not needed here because CMD functions are
	// 	 		called as subfunctions already.

	go func() {
		var (
			start      time.Time
			duration   time.Duration
			err        error
			filesTotal int
			filesRead  int
		)

		filesTotal = len(m.cacheDir.files)

		for _, each := range m.cacheDir.files {
			start = time.Now()

			err = each.Read()
			filesRead += 1

			if err != nil {
				continue
			}

			duration = time.Since(start)

			program.Send(resultMsg{
				duration:   duration,
				err:        err,
				file:       each.path,
				filesRead:  filesRead,
				filesTotal: filesTotal,
				done:       false,
			})
		}

		program.Send(resultMsg{
			done: true,
		})
	}()
}

func initialModel() model {
	t := textinput.NewModel()
	t.Focus()
	t.SetValue(os.Getenv("CACHE_THIS_DIR"))

	s := spinner.NewModel()
	s.Spinner = spinner.Dot
	s.Style = spinnerStyle

	pr := progress.New(progress.WithDefaultGradient())

	results := make([]resultMsg, numDisplayResults)

	m := model{
		done:      false,
		progress:  pr,
		results:   results,
		spinner:   s,
		textInput: t,
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

	return m
}

func StartUI() {
	// if err := tea.NewProgram(initialModel(), tea.WithAltScreen()).Start(); err != nil {
	program = tea.NewProgram(initialModel(), tea.WithAltScreen())
	if err := program.Start(); err != nil {
		fmt.Printf("could not start program: %s\n", err)
		os.Exit(1)
	}
}
