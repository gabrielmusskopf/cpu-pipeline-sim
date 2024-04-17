package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

var debug = true

func Debug(format string, v ...any) {
	if debug {
		if len(v) > 0 {
			log.Printf(format, v)
		} else {
			log.Printf(format)
		}
	}
}

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

func (i Instruction) String() string {
	return fmt.Sprintf("%s", i.Opcode)
}

type Stage struct {
	UserChan        chan rune
	CurrInstruction *Instruction
	IsActive        bool
}

var consts map[string]string
var registers = make([]Register, 0)
var stagesCount = 5
var stages []*Stage

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
	for i, stage := range stages {
		opcode := "EMPTY"
		if stage != nil && stage.CurrInstruction != nil && stage.CurrInstruction.Opcode != "" {
			opcode = string(stage.CurrInstruction.Opcode)
		}
		fmt.Printf("[%d] %s\t", i, opcode)
	}
	fmt.Println()
}

func parseInstruction(line string) *Instruction {
	opcode := strings.Split(line, " ")
	return &Instruction{
		Opcode: Opcode(opcode[0]),
	}
}

func instructionFetch(in chan *Instruction) chan *Instruction {
	s := stages[0]
	out := make(chan *Instruction, 5)
	go func() {
		for instruction := range in {
			Debug("Instruction fetch recieved instruction %v\n", instruction)
			s.CurrInstruction = instruction
			s.IsActive = true
			for {
				select {
				case <-s.UserChan:
					Debug("Instruction fetch toggled")
					s.CurrInstruction = nil
					s.IsActive = false
					out <- instruction
				}
				break
			}
		}
		Debug("Instruction fetch closing output")
		close(out)
	}()

	return out
}

func decodeInstruction(in chan *Instruction) chan *Instruction {
	s := stages[1]
	out := make(chan *Instruction, 5)
	go func() {
		for instruction := range in {
			Debug("Decode instruction recieved instruction %v\n", instruction)
			s.CurrInstruction = instruction
			s.IsActive = true
			for {
				select {
				case <-s.UserChan:
					Debug("Decode instruction toggled")
					s.CurrInstruction = nil
					s.IsActive = false
					out <- instruction
				}
				break
			}
		}
		close(out)
	}()
	return out
}

func executeAddCalc(in chan *Instruction) chan *Instruction {
	s := stages[2]
	out := make(chan *Instruction, 5)
	go func() {
		for instruction := range in {
			Debug("Execute Address Calculation recieved instruction %v\n", instruction)
			s.CurrInstruction = instruction
			s.IsActive = true
			for {
				select {
				case <-s.UserChan:
					Debug("Execute Address Calculation toggled")
					s.CurrInstruction = nil
					s.IsActive = false
					out <- instruction
				}
				break
			}
		}
		close(out)
	}()
	return out
}

func memoryAccess(in chan *Instruction) chan *Instruction {
	s := stages[3]
	out := make(chan *Instruction, 5)
	go func() {
		for instruction := range in {
			Debug("Memory Access recieved instruction %v\n", instruction)
			s.CurrInstruction = instruction
			s.IsActive = true
			for {
				select {
				case <-s.UserChan:
					Debug("Memory Access toggled")
					s.CurrInstruction = nil
					s.IsActive = false
					out <- instruction
				}
				break
			}
		}
		close(out)
	}()
	return out
}

func writeBack(in chan *Instruction) <-chan *Instruction {
	stage := stages[4]
	out := make(chan *Instruction, 5)
	go func() {
		for instruction := range in {
			Debug("Write Back recieved instruction %v\n", instruction)
			stage.CurrInstruction = instruction
			stage.IsActive = true
			for {
				select {
				case <-stage.UserChan:
					Debug("Write Back toggled")
					stage.CurrInstruction = nil
					stage.IsActive = false
					out <- instruction
				}
				break
			}
		}
		close(out)
	}()
	return out
}

func broadcast(v rune) {
	for _, stage := range stages {
		if stage.IsActive {
			stage.UserChan <- v
		}
	}
}

func main() {
	stages = make([]*Stage, 0)
	for i := 0; i < stagesCount; i++ {
		stages = append(stages, &Stage{UserChan: make(chan rune), IsActive: false})
	}

	file, err := os.Open("instrucoes.txt")
	if err != nil {
		log.Fatal(err)
	}

	readConsts(file)
	file.Seek(0, io.SeekStart)

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
			case 'k', 'K':
				broadcast(char)
			case 'h', 'H':
				fmt.Println("Aperte V para ver os estágios, K para avançar o estágio, H para ajuda e Q para sair")
			}
		}
	}()

	instructionsChan := make(chan *Instruction, 512)

	decodeChan := instructionFetch(instructionsChan)
	executeChan := decodeInstruction(decodeChan)
	memAccessChan := executeAddCalc(executeChan)
	writeBackChan := memoryAccess(memAccessChan)
	out := writeBack(writeBackChan)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		instruction := parseInstruction(scanner.Text())
		Debug("Parsed %v instruction\n", instruction)
		instructionsChan <- instruction
	}
	Debug("All instructions sended")
	close(instructionsChan)

	for o := range out {
		Debug("Instruction completed: %v\n", o)
	}
}
