package cli

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Spinner struct {
	message string
	done    chan struct{}
	program *tea.Program
}

type spinnerModel struct {
	spinner spinner.Model
	message string
	done    bool
}

func (m spinnerModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m spinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case doneMsg:
		m.done = true
		return m, tea.Quit
	}
	return m, nil
}

func (m spinnerModel) View() string {
	if m.done {
		return ""
	}
	return fmt.Sprintf("%s %s", m.spinner.View(), StyleMuted.Render(m.message))
}

type doneMsg struct{}

func NewSpinner(message string) *Spinner {
	return &Spinner{
		message: message,
		done:    make(chan struct{}),
	}
}

func (s *Spinner) Start() {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(ColorPrimary)

	model := spinnerModel{
		spinner: sp,
		message: s.message,
	}

	s.program = tea.NewProgram(model)

	go func() {
		s.program.Run()
	}()
}

func (s *Spinner) Stop() {
	if s.program != nil {
		s.program.Send(doneMsg{})
		time.Sleep(50 * time.Millisecond) // Give time for cleanup
	}
}

func (s *Spinner) StopWithSuccess(message string) {
	s.Stop()
	PrintSuccess("%s", message)
}

func (s *Spinner) StopWithError(message string) {
	s.Stop()
	PrintError("%s", message)
}

var loadingActive bool

func PrintLoading(msg string) {
	loadingActive = true
	fmt.Print(StyleMuted.Render(SymbolLoading + " " + msg + "... "))
}

func PrintDone() {
	if loadingActive {
		fmt.Println(StyleSuccess.Render("done"))
		loadingActive = false
	}
}

func PrintFailed() {
	if loadingActive {
		fmt.Println(StyleError.Render("failed"))
		loadingActive = false
	}
}

func WithSpinner[T any](message string, fn func() (T, error)) (T, error) {
	PrintLoading(message)
	result, err := fn()
	if err != nil {
		PrintFailed()
		return result, err
	}
	PrintDone()
	return result, nil
}

func WithSpinnerNoResult(message string, fn func() error) error {
	PrintLoading(message)
	err := fn()
	if err != nil {
		PrintFailed()
		return err
	}
	PrintDone()
	return nil
}
