package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

type Terminal struct {
	Pipeline     Pipeline
	play         bool
	playDuration time.Duration
	playChan     chan rune
	done         chan bool
	ticker       *time.Ticker
}

func help() {
	fmt.Println("Simulador arquitetura pipeline")
	fmt.Println("Instruções:")
	fmt.Println("   v   ver os estágios")
	fmt.Println("   k   avançar o estágio")
	fmt.Println("   p, p <segundos>")
	fmt.Println("       avançar os estágios automaticamente")
	fmt.Println("   d   habilitar/desabilitar logs debug")
	fmt.Println("   h   ajuda")
	fmt.Println("   q   sair")
}

func (t *Terminal) HandleUserInput() {
	go func() {
		inputScan := bufio.NewReader(os.Stdin)
		help()
		for {
			input, err := inputScan.ReadString('\n')
			if err != nil {
				log.Fatal(err)
			}
			parts := strings.Split(input, " ")
			for i, p := range parts {
				parts[i] = strings.TrimSpace(p)
			}

			switch parts[0] {
			case "q", "Q":
				os.Exit(0)
			case "v", "V":
				t.printState()
			case "r", "R":
				t.printRegisters()
			case "d", "D":
				t.toggleDebug()
			case "p", "P":
				t.playDuration = 2 * time.Second
				if len(parts) > 1 {
					d, err := time.ParseDuration(parts[1])
					if err != nil {
						fmt.Printf("ERROR: %s inválido! Formato deve ser valor e sufixo. Usando padrão: 2s\n", parts[1])
					} else {
						t.playDuration = d
					}
				}
				t.togglePlay()
			case "k", "K":
				t.Pipeline.Broadcast('k')
			case "h", "H":
				help()
			}
		}
	}()
}

func (t *Terminal) toggleDebug() {
	debug = !debug
	var debugState string
	if debug {
		debugState = "habilitado"
	} else {
		debugState = "desabilitado"
	}
	fmt.Printf("Debug %s\n", debugState)
}

func (t *Terminal) togglePlay() {
	t.play = !t.play
	if t.play {
		t.ticker = time.NewTicker(t.playDuration)
		t.done = make(chan bool)
		go func() {
			fmt.Printf("Start auto play with %s\n", t.playDuration)
			for {
				select {
				case <-t.done:
					fmt.Println("Done auto play")
					return
				case _ = <-t.ticker.C:
					t.printRegisters()
					t.printState()
					t.Pipeline.Broadcast('k')
				}
			}
		}()
	} else {
		t.ticker.Stop()
		t.done <- true
	}
}

func (t *Terminal) divisionLine() {
	for i := 0; i < 150; i++ {
		fmt.Print("━")
	}
	fmt.Println()
}

func (t *Terminal) printState() {
	t.divisionLine()
	for t, stage := range t.Pipeline.Stages() {
		status := "EMPTY"
		if stage != nil && stage.CurrInstruction != nil && stage.CurrInstruction.Opcode != "" {
			status = stage.CurrInstruction.String()
		}
		if t == 0 { //fetch stage
			status = fmt.Sprintf("PC=%d", stage.CurrPC)
		}

		var active string
		if stage.IsActive {
			active = "A"
		} else {
			active = "I"
		}
		fmt.Printf("(%s)[%s] %s\t\t", active, stage.Nickname, status)
	}
	fmt.Println()
	t.divisionLine()
}

func (t *Terminal) printRegisters() {
	for i := 0; i <= numRegisters; i++ {
		fmt.Print("┌───┐")
	}
	fmt.Println()
	for i := 0; i <= numRegisters; i++ {
		fmt.Printf("│R%02d│", i)
	}
	fmt.Println()
	for i := 0; i <= numRegisters; i++ {
		nick := fmt.Sprintf("R%d", i)
		value := registers[nick]
		fmt.Printf("│%2d │", value)
	}
	fmt.Println()
	for i := 0; i <= numRegisters; i++ {
		fmt.Print("└───┘")
	}
	fmt.Println()
}
