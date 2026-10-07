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

func expand(p string) string {
	if strings.HasPrefix(p, "~") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, p[1:])
	}
	return p
}

func tmux(target, format string) string {
	args := []string{"display-message", "-p"}
	if target != "" {
		args = append(args, "-t", target)
	}
	args = append(args, format)
	out, err := exec.Command("tmux", args...).Output()
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
	sort.Slice(states, func(i, j int) bool {
		wi, wj := 0, 0
		fmt.Sscanf(states[i].Window, "%d", &wi)
		fmt.Sscanf(states[j].Window, "%d", &wj)
		return wi < wj || wi == wj && states[i].Pane < states[j].Pane
	})
	return states
}

type model struct {
	spinner spinner.Model
}

type redrawMsg struct{}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		tea.Tick(time.Second, func(t time.Time) tea.Msg { return redrawMsg{} }),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case redrawMsg:
		return m, tea.Tick(time.Second, func(t time.Time) tea.Msg { return redrawMsg{} })
	}
	return m, nil
}

func (m model) View() string {
	var w, h int
	fmt.Sscanf(tmux(os.Getenv("TMUX_PANE"), "#{pane_width}|#{pane_height}"), "%d|%d", &w, &h)
	usable := w - 1
	if usable <= 0 {
		return ""
	}

	var b strings.Builder
	line := func(s string) {
		b.WriteString(s + strings.Repeat(" ", max(0, usable-len([]rune(ansi.Strip(s))))) + "\n")
	}
	for range 3 { line("") }
	line(titleStyle.Render(" Agents"))
	line("")

	states := readStates()
	if len(states) == 0 {
		line(idleStyle.Render("no agents"))
	}
	for _, s := range states {
		line(m.agentLine(s, usable))
		if s.Harness != "" {
			line(idleStyle.Render("  " + s.Harness))
		}
		line("")
	}

	for rows := 4 + 3*len(states); rows < h && rows < 200; rows++ {
		line("")
	}
	return b.String()
}

func (m model) agentLine(s state, usable int) string {
	var sym string
	if s.Status == "working" {
		sym = workingStyle.Render(" " + m.spinner.View() + " ")
	} else if s.Status == "waiting" {
		sym = waitingStyle.Render(" ? ")
	} else if s.Status == "done" {
		sym = doneStyle.Render(" ✓ ")
	} else {
		sym = idleStyle.Render(" · ")
	}
	num := s.Window
	if i := strings.Index(num, ":"); i >= 0 {
		num = num[:i]
	}
	if num == "" {
		num = "?"
	}
	title := tmux(s.Pane, "#{window_name}")
	if title == "" {
		title = "?"
	}
	title = strings.NewReplacer("\n", " ", "\r", " ").Replace(title)
	budget := usable - len([]rune(ansi.Strip(sym+" "+num+": "))) - 2
	if budget < 4 {
		budget = 4
	}
	if len([]rune(title)) > budget {
		title = ansi.Truncate(title, budget-1, "…")
	}
	return sym + " " + num + ": " + title
}

func main() {
	sp := spinner.New(spinner.WithSpinner(spinner.MiniDot))
	p := tea.NewProgram(model{spinner: sp})
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "sidebar:", err)
		os.Exit(1)
	}
}
