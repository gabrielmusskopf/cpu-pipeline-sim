package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

type Register = [8]int

type Opcode string

func (o Opcode) String() string {
	return string(o)
}

const (
	ADD  Opcode = "add"
	ADDI Opcode = "addi"
	SUB  Opcode = "sub"
	SUBI Opcode = "subi"
	BEQ  Opcode = "beq"
	J    Opcode = "j"
)

type Instruction struct {
	Opcode Opcode
	Op1    string
	Op2    string
	Op3    string
	Temp1  string
	Temp2  string
	Temp3  string
	Valid  bool
}

var consts map[string]string
var registers = make([]Register, 0)
var states = make([]*Instruction, 5)
var userChan = make(chan rune)

func readConsts(file *os.File) {
	scan := bufio.NewScanner(file)

	constsStarted := false
	consts = make(map[string]string)

	for scan.Scan() {
		line := scan.Text()
		if strings.Contains(line, "halt") {
			constsStarted = true
		}
		if constsStarted {
			lineParts := strings.Split(line, " ")
			consts[lineParts[0]] = lineParts[len(lineParts)-1]
		}
	}
}

func printState() {
	for i, state := range states {
		opcode := "EMPTY"
		if state.Opcode != "" {
			opcode = string(state.Opcode)
		}
		fmt.Printf("[%d] %s\t", i, opcode)
	}
	fmt.Println()
}

func cycle() {
}

func parseInstruction(line string) *Instruction {
	opcode := strings.Split(line, " ")
	return &Instruction{
		Opcode: Opcode(opcode[0]),
	}
}

func instructionFetch(in chan *Instruction) chan *Instruction {
	out := make(chan *Instruction)

	go func() {
		for instruction := range in {
			fmt.Printf("Fetch recebeu uma instrução %v\n", instruction.Opcode)
			states[0] = instruction
			select {
			case userInput := <-userChan:
				switch userInput {
				case 'q', 'Q':
					os.Exit(0)
				case 'k', 'K':
					states[0] = nil
					out <- instruction
				}
			}
			out <- instruction
		}
		close(out)
	}()

	return out
}

func decodeInstruction(in chan *Instruction) chan *Instruction {
	out := make(chan *Instruction)
	go func() {
		for instruction := range in {
			fmt.Printf("Decode recebeu uma instrução: %v\n", instruction.Opcode)
			states[1] = instruction
			select {
			case userInput := <-userChan:
				switch userInput {
				case 'q', 'Q':
					os.Exit(0)
				case 'k', 'K':
					states[1] = nil
					out <- instruction
				}
			}
		}
		close(out)
	}()
	return out
}

func executeAddCalc(in chan *Instruction) chan *Instruction {
	out := make(chan *Instruction)
	go func() {
		for instruction := range in {
			fmt.Printf("Execute Address Calc recebeu uma instrução: %v\n", instruction.Opcode)
			states[2] = instruction
			select {
			case userInput := <-userChan:
				switch userInput {
				case 'q', 'Q':
					os.Exit(0)
				case 'k', 'K':
					states[2] = nil
					out <- instruction
				}
			}
		}
		close(out)
	}()
	return out
}

func memoryAccess(in chan *Instruction) chan *Instruction {
	out := make(chan *Instruction)
	go func() {
		for instruction := range in {
			fmt.Printf("Memory Access recebeu uma instrução: %v\n", instruction.Opcode)
			states[3] = instruction
			select {
			case userInput := <-userChan:
				switch userInput {
				case 'q', 'Q':
					os.Exit(0)
				case 'k', 'K':
					states[3] = nil
					out <- instruction
				}
			}
		}
		close(out)
	}()
	return out
}

func writeBack(in chan *Instruction) chan *Instruction {
	out := make(chan *Instruction)
	go func() {
		for instruction := range in {
			fmt.Printf("Write Back recebeu uma instrução: %v\n", instruction.Opcode)
			states[4] = instruction
			select {
			case userInput := <-userChan:
				switch userInput {
				case 'q', 'Q':
					os.Exit(0)
				case 'k', 'K':
					states[4] = nil
					out <- instruction
				}
			}
		}
		close(out)
	}()
	return out
}

func main() {
	file, err := os.Open("instrucoes.txt")
	if err != nil {
		log.Fatal(err)
	}

	readConsts(file)
	file.Seek(0, io.SeekStart)

	instructionsChan := make(chan *Instruction)

	decodeChan := instructionFetch(instructionsChan)
	executeChan := decodeInstruction(decodeChan)
	memAccessChan := executeAddCalc(executeChan)
	writeBackChan := memoryAccess(memAccessChan)
	out := writeBack(writeBackChan)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		instructionsChan <- parseInstruction(scanner.Text())
	}

	fmt.Printf("Instrução finalizada %v\n", <-out)
	fmt.Printf("Instrução finalizada %v\n", <-out)

    inputScan := bufio.NewReader(os.Stdin)

    fmt.Println("Aperte K para avançar o estágio e Q para sair")
    for {
        char, _, err := inputScan.ReadRune()
        if err != nil {
            log.Fatal(err)
        }
        userChan<-char
    }

}
