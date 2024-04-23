package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

var pipeline Pipeline

type toggleStagesCmd int

type model struct {
	sub           chan interface{}
	quitting      bool
	stages        []*stage
	messages      []string
	registers     map[string]int8
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
		m.messages = append(m.messages, msg.message)

	case toggleStagesMsg:
		pipeline.Broadcast('k')

	case tea.KeyMsg:
		key := msg.String()

		if m.askParams {
			switch key {
			case "ctrl+c":
				return m, quit

			case "enter":
				m.askParams = false
				v := m.input.Value()
				m.input.Reset()

				duration, err := time.ParseDuration(v)
				if err != nil {
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

		switch key {
		case "ctrl+c", "q":
			return m, quit

		case "k":
			return m, toggleStages

		case "p", "P":
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

	return m, waitForActivity(m.sub)
}

func (m model) View() string {
	var sb strings.Builder

	sb.WriteString("Simulador pipeline MIPS\n\n")

	sb.WriteString(fmt.Sprintf("Autoplay: %v", m.autoplay))
	if m.autoplay {
		sb.WriteString(fmt.Sprintf(" [%v]", m.autoplayDelay))
	}
	sb.WriteString("\n\n")

	for i := 0; i < len(m.registers); i++ {
		name := fmt.Sprintf("R%d", i)
		sb.WriteString(fmt.Sprintf("%s=%d\t", name, m.registers[name]))

		if i == (len(m.registers)-1)/2 {
			sb.WriteString("\n")
		}
	}
	sb.WriteString("\n\n")

	for _, stage := range m.stages {
		s := fmt.Sprintf("[%s] %v\t\t", stage.nickname, stage.value)
		sb.WriteString(s)
	}
	sb.WriteString("\n\n")

	if m.askParams {
		sb.WriteString(m.input.View())
		sb.WriteString("\n\n")
	}

	sb.WriteString("Eventos\n")
	latest := m.messages
	if len(m.messages) >= 15 {
		latest = m.messages[len(m.messages)-15:]
	}
	for _, message := range latest {
		sb.WriteString(message)
	}

	if m.quitting {
		sb.WriteString("Finishing!\n")
		sb.WriteString("\n")
	}

	return sb.String()
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

	return model{
		sub:          events,
		stages:       stages,
		registers:    registers,
		input:        ti,
		autoplayDone: make(chan bool),
	}
}

func RunCmd(pipe Pipeline, regs map[string]int8, events chan interface{}) {

	p := tea.NewProgram(initModel(pipe, regs))

	if _, err := p.Run(); err != nil {
		fmt.Println("could not start program:", err)
		os.Exit(1)
	}
}
