package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

var (
	titleStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		return lipgloss.NewStyle().BorderStyle(b).Padding(0, 1)
	}()

	infoStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Left = "┤"
		return titleStyle.Copy().BorderStyle(b)
	}()
)

var pipeline Pipeline

// keyMap defines a set of keybindings. To work for help it must satisfy
// key.Map. It could also very easily be a map[string]key.Binding.
type keyMap struct {
	L    key.Binding
	K    key.Binding
	J    key.Binding
	D    key.Binding
	P    key.Binding
	Help key.Binding
	Quit key.Binding
}

// ShortHelp returns keybindings to be shown in the mini help view. It's part
// of the key.Map interface.
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

// FullHelp returns keybindings for the expanded help view. It's part of the
// key.Map interface.
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.J, k.K, k.L, k.P, k.D},
		{k.Help, k.Quit},
	}
}

var keys = keyMap{
	P: key.NewBinding(
		key.WithKeys("p", "P"),
		key.WithHelp("p/P", "autoplay"),
	),
	D: key.NewBinding(
		key.WithKeys("d", "D"),
		key.WithHelp("d/D", "toggle debug"),
	),
	J: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "scroll events down"),
	),
	K: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "scroll events up"),
	),
	L: key.NewBinding(
		key.WithKeys("l", "L"),
		key.WithHelp("l/L", "toggle stages"),
	),
	Help: key.NewBinding(
		key.WithKeys("?", "h", "H"),
		key.WithHelp("?/h", "toggle help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "esc", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

type model struct {
	sub           chan interface{}
	quitting      bool
	stages        []*stage
	messages      []string
	messagesView  viewport.Model
	messagesStyle lipgloss.Style
	registers     map[string]int8
	keys          keyMap
	help          help.Model
	askParams     bool
	input         textinput.Model
	autoplay      bool
	autoplayDelay time.Duration
	autoplayDone  chan bool
}

type stage struct {
	nickname string
	name     string
	value    any
}

func initModel(pipe Pipeline, regs map[string]int8) model {
	pipeline = pipe

	registers := make(map[string]int8)
	for k, v := range regs {
		registers[k] = v
	}

	stages := make([]*stage, 0)
	for _, s := range pipeline.Stages() {
		stages = append(stages, &stage{
			name:     s.Name,
			nickname: s.Nickname,
		})
	}

	ti := textinput.New()
	ti.CharLimit = 5
	ti.Width = 20

	vp := viewport.New(150, 15)
	vp.SetContent("Messages")

	//p.ActiveDot = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "235", Dark: "252"}).Render("•")
	//p.InactiveDot = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "250", Dark: "238"}).Render("•")

	return model{
		sub:           events,
		stages:        stages,
		registers:     registers,
		input:         ti,
		autoplayDone:  make(chan bool),
		keys:          keys,
		help:          help.New(),
		messagesView:  vp,
		messagesStyle: lipgloss.NewStyle().Background(lipgloss.Color("#7D56F4")),
	}
}

func waitForActivity(sub chan interface{}) tea.Cmd {
	return func() tea.Msg {
		return <-sub
	}
}

func autoplayStages(m model) tea.Cmd {
	return func() tea.Msg {
		for {
			select {
			case <-m.autoplayDone:
				return responseMsg{}
			default:
				m.sub <- toggleStagesMsg{}
				time.Sleep(m.autoplayDelay)
			}
		}
	}
}

func toggleStages() tea.Msg {
	return toggleStagesMsg{}
}

func quit() tea.Msg {
	return quitMsg{}
}

func (m model) Init() tea.Cmd {
	return waitForActivity(m.sub)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {

	case quitMsg:
		m.quitting = true
		return m, tea.Quit

	case responseMsg:

	case stageToggledMsg:
		s := m.stages[msg.position]
		s.value = msg.value

	case registerUpdatedMsg:
		m.registers[msg.name] = msg.value

	case debugMsg:
		m.messages = append([]string{msg.message}, m.messages...)
		m.messagesView.SetContent(strings.Join(m.messages, ""))

	case toggleStagesMsg:
		pipeline.Broadcast('k') //TODO: Alterar para bool ou struct{}

	case tea.WindowSizeMsg:
		m.help.Width = msg.Width

	case tea.KeyMsg:
		if m.askParams {
			switch msg.String() {
			case "ctrl+c":
				return m, quit

			case "enter":
				m.askParams = false
				v := m.input.Value()
				m.input.Reset()

				duration, err := time.ParseDuration(v)
				if err != nil {
					Debug("Invalid '%v' duration. Using default\n", v)
					duration = 2 * time.Second
				}
				m.autoplay = true
				m.autoplayDelay = duration
				Debug("Activating autoplay mode\n")

				return m, autoplayStages(m)
			}

			m.input, cmd = m.input.Update(msg)
			return m, cmd
		}

		switch {
		//case "ctrl+c", "q", "Q":
		case key.Matches(msg, m.keys.Quit):
			return m, quit

		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll

		//case "d", "D":
		case key.Matches(msg, m.keys.D):
			debug = !debug
			Info("Debug: %v\n", debug)

		//case "l":
		case key.Matches(msg, m.keys.L):
			return m, toggleStages

		//case "p", "P":
		case key.Matches(msg, m.keys.P):
			if m.autoplay {
				m.autoplayDone <- true
				m.autoplay = false
				Debug("Dectivating autoplay mode\n")
			} else {
				m.askParams = true
				m.input.Placeholder = "Duration"
				m.input.Focus()
			}
		}
	}

	m.messagesView, cmd = m.messagesView.Update(msg)

	return m, tea.Batch(cmd, waitForActivity(m.sub))
}

func (m model) View() string {
	var sb strings.Builder

	sb.WriteString("Simulador pipeline MIPS\n\n")

	// Informações gerais
	sb.WriteString(fmt.Sprintf("Autoplay: %v ", m.autoplay))
	if m.autoplay {
		sb.WriteString(fmt.Sprintf("[%v] ", m.autoplayDelay))
	}

	sb.WriteString(fmt.Sprintf("\nDebug:\t  %v", debug))
	sb.WriteString("\n\n")

	// Registradores
	for i := 0; i < len(m.registers); i++ {
		name := fmt.Sprintf("R%d", i)
		sb.WriteString(fmt.Sprintf("%s=%d\t", name, m.registers[name]))

		if i == (len(m.registers)-1)/2 {
			sb.WriteString("\n")
		}
	}
	sb.WriteString("\n\n")

	// Estágios
	for _, stage := range m.stages {
		s := fmt.Sprintf("[%s] %v\t\t", stage.nickname, stage.value)
		sb.WriteString(s)
	}
	sb.WriteString("\n\n")

	// Input parâmetros
	if m.askParams {
		sb.WriteString(m.input.View())
		sb.WriteString("\n\n")
	}

	// Eventos
	sb.WriteString(m.headerView() + "\n")
	sb.WriteString(m.messagesView.View())
	sb.WriteString(m.footerView() + "\n")
	sb.WriteString("\n")

	if m.quitting {
		sb.WriteString("Finishing!\n\n")
		return sb.String()
	}

	sb.WriteString(m.help.View(m.keys))
	sb.WriteString("\n")

	return sb.String()
}

func (m model) headerView() string {
	//title := titleStyle.Render("Eventos")
	title := "── Eventos "
	line := strings.Repeat("─", max(0, m.messagesView.Width-lipgloss.Width(title)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}

func (m model) footerView() string {
	info := infoStyle.Render(fmt.Sprintf("%3.f%%", m.messagesView.ScrollPercent()*100))
	line := strings.Repeat("─", max(0, m.messagesView.Width-lipgloss.Width(info)))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}

func RunCmd(pipe Pipeline, regs map[string]int8, events chan interface{}) {

	p := tea.NewProgram(initModel(pipe, regs))

	if _, err := p.Run(); err != nil {
		fmt.Println("could not start program:", err)
		os.Exit(1)
	}
}
