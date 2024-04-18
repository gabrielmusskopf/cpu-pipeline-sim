package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"time"
)

type Terminal struct {
	AutoPlay bool
	playChan chan rune
	done     chan bool
	ticker   *time.Ticker
}

func (t *Terminal) HandleUserInput() {
	go func() {
		inputScan := bufio.NewReader(os.Stdin)

		fmt.Println("Simulador arquitetura pipeline")
		fmt.Println("Instruções:")
		fmt.Println("   v   ver os estágios")
		fmt.Println("   k   avançar o estágio")
		fmt.Println("   p   avançar os estágios automaticamente")
		fmt.Println("   d   habilitar/desabilitar logs debug")
		fmt.Println("   h   ajuda")
		fmt.Println("   q   sair")
		for {
			char, _, err := inputScan.ReadRune()
			if err != nil {
				log.Fatal(err)
			}
			switch char {
			case 'q', 'Q':
				os.Exit(0)
			case 'v', 'V':
				printState()
			case 'r', 'R':
				printRegisters()
			case 'd', 'D':
				debug = !debug
				var debugState string
				if debug {
					debugState = "habilitado"
				} else {
					debugState = "desabilitado"
				}
				fmt.Printf("Debug %s\n", debugState)
			case 'p', 'P':
				t.togglePlay()
			case 'k', 'K':
				Broadcast(char)
			case 'h', 'H':
				fmt.Println("Aperte V para ver os estágios, K para avançar o estágio, H para ajuda e Q para sair")
			}
		}
	}()
}

func (t *Terminal) togglePlay() {
	t.AutoPlay = !t.AutoPlay
	if t.AutoPlay {
		t.ticker = time.NewTicker(2 * time.Second)
		t.done = make(chan bool)
		go func() {
			fmt.Println("Start auto play")
			for {
				select {
				case <-t.done:
					fmt.Println("Done auto play")
					return
				case _ = <-t.ticker.C:
					printRegisters()
					printState()
					Broadcast('k')
				}
			}
		}()
	} else {
		t.ticker.Stop()
		t.done <- true
	}
}

func divisionLine() {
	for i := 0; i < 150; i++ {
		fmt.Print("━")
	}
	fmt.Println()
}

func printState() {
	divisionLine()
	for t, stage := range stages {
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
	divisionLine()
}

func printRegisters() {
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
