package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

type Terminal struct {
}

func (t *Terminal) HandleUserInput() {
	go func() {
		inputScan := bufio.NewReader(os.Stdin)

		fmt.Println("Simulador arquitetura pipeline")
		fmt.Println("Aperte V para ver os estágios, K para avançar o estágio, H para ajuda e Q para sair")
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
			case 'k', 'K':
				Broadcast(char)
			case 'h', 'H':
				fmt.Println("Aperte V para ver os estágios, K para avançar o estágio, H para ajuda e Q para sair")
			}
		}
	}()
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
		fmt.Print("╭───╮ ")
	}
	fmt.Println()
	for i := 0; i <= numRegisters; i++ {
		fmt.Printf("│R%02d│ ", i)
	}
	fmt.Println()
	for i := 0; i <= numRegisters; i++ {
		nick := fmt.Sprintf("R%d", i)
		value := registers[nick]
		fmt.Printf("│%2d │ ", value)
	}
	fmt.Println()
	for i := 0; i <= numRegisters; i++ {
		fmt.Print("╰───╯ ")
	}
	fmt.Println()
}
