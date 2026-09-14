// Command sidebar renders the agent sidebar canvas: one line per agent
// pane, bubbles MiniDot spinner for working agents, cell-perfect truncation.
//
// Runs as a bubbletea program in the tmux pane; standard (non-altscreen)
// renderer, full-frame View, spinner ticks drive repaint.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

type state struct {
	Pane    string `json:"pane"`
	Status  string `json:"status"`
	Summary string `json:"summary"`
	Window  string `json:"window"`
	Harness string `json:"harness"`
}

var (
	titleStyle   = lipgloss.NewStyle().Bold(true)
	workingStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	waitingStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	doneStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	idleStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)

func styleFor(s string) lipgloss.Style {
	switch s {
	case "working":
		return workingStyle
	case "waiting":
		return waitingStyle
	case "done":
		return doneStyle
	default:
		return idleStyle
	}
}

func symFor(s string) string {
	switch s {
	case "waiting":
		return "?"
	case "done":
		return "✓"
	default:
		return "·"
	}
}

func expand(p string) string {
	if strings.HasPrefix(p, "~") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, p[1:])
	}
	return p
}

func tmux(format string) string {
	args := []string{"display-message", "-p"}
	if pane := os.Getenv("TMUX_PANE"); pane != "" {
		args = append(args, "-t", pane)
	}
	args = append(args, format)
	out, err := exec.Command("tmux", args...).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func paneFormat(pane, format string) string {
	out, err := exec.Command("tmux", "display-message", "-p", "-t", pane, format).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func readStates() []state {
	dir := expand("~/.cache/tmux-agent-sidebar")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	states := []state{}
	for _, e := range entries {
		if filepath.Ext(e.Name()) != ".json" {
			continue
		}
		b, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		var s state
		if json.Unmarshal(b, &s) == nil && s.Pane != "" {
			states = append(states, s)
		}
	}
	sort.Slice(states, func(i, j int) bool { return states[i].Window < states[j].Window })
	return states
}

func visibleLen(s string) int {
	return len([]rune(ansi.Strip(s)))
}

type model struct {
	spinner spinner.Model
}

type redrawMsg struct{}

func redraw() tea.Cmd { return tea.Tick(time.Second, func(t time.Time) tea.Msg { return redrawMsg{} }) }

func (m model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, redraw())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case redrawMsg:
		return m, redraw()
	}
	return m, nil
}

func (m model) View() string {
	var w, h int
	fmt.Sscanf(tmux("#{pane_width}|#{pane_height}"), "%d|%d", &w, &h)
	usable := w - 1
	if usable <= 0 {
		return ""
	}

	var b strings.Builder
	line := func(s string) {
		b.WriteString(s)
		b.WriteString(strings.Repeat(" ", max(0, usable-visibleLen(s))))
		b.WriteString("\n")
	}
	line("")
	line("")
	line("")
	line(titleStyle.Render(" Agents"))
	line("")

	states := readStates()
	agents := 0
	for _, s := range states {
		line(m.agentLine(s, usable))
		if s.Harness != "" {
			line(idleStyle.Render("  " + s.Harness))
		}
		line("")
		agents++
	}
	if agents == 0 {
		line(idleStyle.Render("no agents"))
	}

	// blank rows below the frame, capped at pane height
	rows := 4 + 3*agents
	for rows < h && rows < 200 {
		line("")
		rows++
	}
	return b.String()
}

func (m model) agentLine(s state, usable int) string {
	var sym string
	if s.Status == "working" {
		sym = workingStyle.Render(" " + m.spinner.View() + " ")
	} else {
		sym = styleFor(s.Status).Render(" " + symFor(s.Status) + " ")
	}
	num := s.Window
	if i := strings.Index(num, ":"); i >= 0 {
		num = num[:i]
	}
	if num == "" {
		num = "?"
	}
	title := paneFormat(s.Pane, "#{window_name}")
	if title == "" {
		title = "?"
	}
	title = strings.NewReplacer("\n", " ", "\r", " ").Replace(title)
	budget := usable - visibleLen(sym+" "+num+": ") - 2
	if budget < 4 {
		budget = 4
	}
	if len([]rune(title)) > budget {
		title = ansi.Truncate(title, budget-1, "…")
	}
	return sym + " " + num + ": " + title
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func main() {
	sp := spinner.New(spinner.WithSpinner(spinner.MiniDot))
	p := tea.NewProgram(model{spinner: sp})
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "sidebar:", err)
		os.Exit(1)
	}
}
