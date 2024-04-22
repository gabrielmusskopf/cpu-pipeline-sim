package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

var pipeline Pipeline

type model struct {
	sub       chan interface{}
	quitting  bool
	stages    []*stage
	messages  []string
	registers map[string]int8
}

type stage struct {
	nickname string
	name     string
	value    any
}

type responseMsg struct{}
type quitMsg struct{
cause string
}

type stageToggledMsg struct {
	position int
	value    any
}

type registerUpdatedMsg struct {
	name  string
	value int8
}

type debugMsg struct {
	message string
}

func waitForActivity(sub chan interface{}) tea.Cmd {
	return func() tea.Msg {
		e := <-sub
		return e
	}
}

func (m model) Init() tea.Cmd {
	return waitForActivity(m.sub)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case quitMsg:
		m.quitting = true
		return m, tea.Quit

	case responseMsg:
		return m, waitForActivity(m.sub)

	case stageToggledMsg:
		s := m.stages[msg.position]
		s.value = msg.value
		return m, waitForActivity(m.sub)

	case registerUpdatedMsg:
		m.registers[msg.name] = msg.value
		return m, waitForActivity(m.sub)

	case debugMsg:
		m.messages = append(m.messages, msg.message)
		return m, waitForActivity(m.sub)

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit

		case "k":
			pipeline.Broadcast('k')

			return m, waitForActivity(m.sub)
		}
	}

	return m, nil
}

func (m model) View() string {
	var sb strings.Builder

	sb.WriteString("Simulador pipeline MIPS\n\n")

	for i := 0; i < len(m.registers); i++ {
		name := fmt.Sprintf("R%d", i)
		sb.WriteString(fmt.Sprintf("%s=%d  ", name, m.registers[name]))
	}
	sb.WriteString("\n\n")

	for _, stage := range m.stages {
		s := fmt.Sprintf("[%s] %v\t\t", stage.nickname, stage.value)
		sb.WriteString(s)
	}
	sb.WriteString("\n\n")

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

func RunCmd(pipe Pipeline, regs map[string]int8, events chan interface{}) {
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
	p := tea.NewProgram(model{
		sub:       events,
		stages:    stages,
		registers: registers,
	})

	if _, err := p.Run(); err != nil {
		fmt.Println("could not start program:", err)
		os.Exit(1)
	}
}
