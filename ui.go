package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

type venvItem struct {
	v Venv
}

func (i venvItem) Title() string { return "🐍 " + i.v.Alias }
func (i venvItem) Description() string {
	ver := pythonVersion(i.v.Path)
	cmds := ""
	if len(i.v.Commands) > 0 {
		names := make([]string, 0, len(i.v.Commands))
		for k := range i.v.Commands {
			names = append(names, k)
		}
		cmds = "  ⚡ " + strings.Join(names, ", ")
	}
	return fmt.Sprintf("Python %s  📁 %s%s", ver, i.v.Path, cmds)
}
func (i venvItem) FilterValue() string { return i.v.Alias }

type uiAction int

const (
	actNone uiAction = iota
	actShell
	actRun
	actExec
	actRemove
)

type uiModel struct {
	list     list.Model
	chosen   *Venv
	action   uiAction
	quitting bool
	help     string
}

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true).
			Padding(0, 1)
	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Padding(0, 2)
)

func (m uiModel) Init() tea.Cmd { return nil }

func (m uiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width-4, msg.Height-6)
	case tea.KeyMsg:
		if m.list.FilterState() == list.Filtering {
			break
		}
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		case "enter", "s":
			if it, ok := m.list.SelectedItem().(venvItem); ok {
				m.chosen = &it.v
				m.action = actShell
				return m, tea.Quit
			}
		case "r":
			if it, ok := m.list.SelectedItem().(venvItem); ok {
				m.chosen = &it.v
				m.action = actRun
				return m, tea.Quit
			}
		case "x":
			if it, ok := m.list.SelectedItem().(venvItem); ok {
				m.chosen = &it.v
				m.action = actExec
				return m, tea.Quit
			}
		case "d":
			if it, ok := m.list.SelectedItem().(venvItem); ok {
				m.chosen = &it.v
				m.action = actRemove
				return m, tea.Quit
			}
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m uiModel) View() string {
	if m.quitting && m.chosen == nil {
		return ""
	}
	return "\n" + titleStyle.Render("🐍 pvm — Python Venv Manager") + "\n" +
		m.list.View() + "\n" +
		helpStyle.Render("enter/s: shell  r: run  x: exec  d: remove  /: filter  q: quit")
}

func runUI() error {
	c := mustConfig()
	if len(c.Venvs) == 0 {
		fmt.Println("No venvs registered. Use `pvm scan <path>` first.")
		return nil
	}
	items := make([]list.Item, 0, len(c.Venvs))
	for _, v := range c.Venvs {
		items = append(items, venvItem{v: v})
	}
	del := list.NewDefaultDelegate()
	del.Styles.SelectedTitle = del.Styles.SelectedTitle.Foreground(lipgloss.Color("205")).BorderForeground(lipgloss.Color("205"))
	del.Styles.SelectedDesc = del.Styles.SelectedDesc.Foreground(lipgloss.Color("170")).BorderForeground(lipgloss.Color("205"))

	l := list.New(items, del, 80, 20)
	l.Title = fmt.Sprintf("%d venv(s)", len(items))
	l.SetShowHelp(false)
	l.SetShowStatusBar(true)

	m := uiModel{list: l}
	result, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	if err != nil {
		return err
	}
	final := result.(uiModel)
	if final.chosen == nil {
		return nil
	}
	return handleAction(final.chosen, final.action)
}

func handleAction(v *Venv, act uiAction) error {
	switch act {
	case actShell:
		fmt.Printf("Opening shell for %s...\n", v.Alias)
		c, err := shellCommand(v.Path)
		if err != nil {
			return err
		}
		c.Stdin, c.Stdout, c.Stderr = os.Stdin, os.Stdout, os.Stderr
		return c.Run()
	case actRun:
		fmt.Printf("Python args for %s (e.g. `script.py` or `-m pip list`): ", v.Alias)
		reader := bufio.NewReader(os.Stdin)
		line, _ := reader.ReadString('\n')
		parts := strings.Fields(strings.TrimSpace(line))
		c := exec.Command(pythonExe(v.Path), parts...)
		c.Env = activatedEnv(v.Path)
		c.Stdin, c.Stdout, c.Stderr = os.Stdin, os.Stdout, os.Stderr
		return c.Run()
	case actExec:
		fmt.Printf("Command for %s: ", v.Alias)
		reader := bufio.NewReader(os.Stdin)
		line, _ := reader.ReadString('\n')
		c, err := commandFromString(line, v.Path)
		if err != nil {
			return err
		}
		if c == nil {
			return nil
		}
		c.Stdin, c.Stdout, c.Stderr = os.Stdin, os.Stdout, os.Stderr
		return c.Run()
	case actRemove:
		fmt.Printf("Remove %s? (y/N): ", v.Alias)
		var ans string
		fmt.Scanln(&ans)
		if strings.ToLower(ans) != "y" {
			return nil
		}
		cfg := mustConfig()
		_, i := cfg.Find(v.Alias)
		if i >= 0 {
			cfg.Venvs = append(cfg.Venvs[:i], cfg.Venvs[i+1:]...)
			if err := cfg.Save(); err != nil {
				return err
			}
			fmt.Printf("Removed %s\n", v.Alias)
		}
	}
	return nil
}

func uiCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ui",
		Short: "Interactive TUI to browse and run venvs",
		Run: func(cmd *cobra.Command, args []string) {
			if err := runUI(); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		},
	}
}
