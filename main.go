package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

var debug = true

func Debug(format string, v ...any) {
	if debug {
		log.Printf(format, v...)
	}
}

var pipeline *Pipeline

type Pipeline struct {
	File   []byte
	PC     int
	Labels map[string]int // Label: PC
	Lines  int
}

func NewPipeline() *Pipeline {
	file, err := os.Open("instrucoes.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	b, err := io.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}
	pipeline := &Pipeline{
		File: b,
		PC:   0,
	}

	pipeline.ParseFile()

	return pipeline
}

func (p *Pipeline) ReadLine(num int) string {
	reader := bytes.NewReader(p.File)
	sc := bufio.NewScanner(reader)
	for i := 1; i <= num; i++ {
		sc.Scan()
	}
	return sc.Text()
}

func (p *Pipeline) ParseFile() {
	lines := 0
	labels := make(map[string]int)

	reader := bytes.NewReader(p.File)
	scan := bufio.NewScanner(reader)
	for scan.Scan() {
		lines++
		lineParts := strings.Split(scan.Text(), " ")
		key := lineParts[0]

		if !IsOpcode(key) {
			labels[key] = lines
			Debug("Parsed [%s: %d] constant\n", key, labels[key])
		}
	}

	p.Labels = labels
	p.Lines = lines
}

var numRegisters = 32
var registers map[string]int8

type Opcode string

const (
	ADD  Opcode = "add"
	ADDI Opcode = "addi"
	SUB  Opcode = "sub"
	SUBI Opcode = "subi"
	BEQ  Opcode = "beq"
	J    Opcode = "j"
	NOOP Opcode = "noop"
)

func (o Opcode) String() string {
	return string(o)
}

func IsOpcode(s string) bool {
	o := Opcode(s)
	return ADD == o ||
		ADDI == o ||
		SUB == o ||
		SUBI == o ||
		BEQ == o ||
		J == o ||
		NOOP == o
}

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
	var sb strings.Builder
	sb.WriteString(i.Opcode.String())

	if len(i.Op1) != 0 {
		sb.WriteString(" ")
		sb.WriteString(i.Op1)
	}
	if len(i.Op2) != 0 {
		sb.WriteString(" ")
		sb.WriteString(i.Op2)
	}
	if len(i.Op3) != 0 {
		sb.WriteString(" ")
		sb.WriteString(i.Op3)
	}

	return sb.String()
}

// Substiuindo lw: addi R0 R1 -1 = Soma R0 com neg1 e coloca no R1
func AddiOperation(i *Instruction) error {
	op1, ok := registers[i.Op1]
	if !ok {
		i.Valid = false
		return fmt.Errorf("Register %s does not exist", i.Op1)
	}
	_, ok = registers[i.Op2]
	if !ok {
		i.Valid = false
		return fmt.Errorf("Register %s does not exist", i.Op2)
	}
	op3 := registers[i.Op3]
	pc, ok := pipeline.Labels[i.Op3]
	if ok {
		// Contains a label. "Jump" to related PC to get the int8 value
		op3 = spy(pc)
	}
	registers[i.Op2] = op1 + op3
	return nil
}

// "Jump" to PC and get the value
func spy(pc int) int8 {
	line := pipeline.ReadLine(pc)
	parts := strings.Split(line, " ")
	r, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		fmt.Printf("ERROR: R3 is not a number. Maybe it was not decoded properly")
		return 0
	}
	return int8(r)
}

// add R0 R1 R2
// R0 = R1 + R2
func AddOperation(i *Instruction) {
	op1Nick := fmt.Sprintf("R%s", i.Op1)
	op2Nick := fmt.Sprintf("R%s", i.Op2)
	op3Nick := fmt.Sprintf("R%s", i.Op3)

	_, ok := registers[op1Nick]
	if !ok {
		i.Valid = false
		fmt.Printf("ERROR: Register %s does not exist\n", i.Op1)
	}
	op2, ok := registers[op2Nick]
	if !ok {
		i.Valid = false
		fmt.Printf("ERROR: Register %s does not exist\n", i.Op1)
	}
	op3, ok := registers[op3Nick]
	if !ok {
		i.Valid = false
		fmt.Printf("ERROR: Register %s does not exist\n", i.Op1)
	}

	registers[op1Nick] = op2 + op3
}

type Stage struct {
	UserChan        chan rune
	CurrInstruction *Instruction
	CurrPC          int
	IsActive        bool
}

var stages [5]*Stage

func divisionLine() {
	for i := 0; i < 100; i++ {
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

		fmt.Printf("[%d] %s\t", t, status)
	}
	fmt.Println()
	divisionLine()
}

func printRegisters() {
	nickFormat := "│R%02d│ "

	for i := 0; i <= numRegisters; i++ {
		fmt.Print("╭───╮ ")
	}
	fmt.Println()
	for i := 0; i <= numRegisters; i++ {
		fmt.Printf(nickFormat, i)
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

// in Program counter (PC)
func instructionFetch(in chan int) chan string {
	s := stages[0]
	out := make(chan string, 5)
	go func() {
		Debug("Fetch goroutine started and is waiting for messages\n")
		for pc := range in {
			Debug("Instruction fetch recieved PC %d\n", pc)
			s.CurrPC = pc
			s.IsActive = true
			instruction := pipeline.ReadLine(pc)
			for {
				select {
				case <-s.UserChan:
					out <- instruction
					s.IsActive = false
				}
				break
			}
		}
		Debug("Fetch will not recieve anything else\n")
		close(out)
	}()

	return out
}

// in Raw instrucion line channel
func decodeInstruction(in chan string) chan *Instruction {
	s := stages[1]
	out := make(chan *Instruction, 5)
	go func() {
		Debug("Decode goroutine started and is waiting for messages\n")
		for raw := range in {
			Debug("Decode instruction recieved instruction %s\n", raw)
			instruction := parseInstruction(raw)
			s.CurrInstruction = instruction
			s.IsActive = true
			for {
				select {
				case <-s.UserChan:
					out <- instruction
					s.CurrInstruction = nil
					s.IsActive = false
				}
				break
			}
		}
		Debug("Decode will not recieve anything else\n")
		close(out)
	}()
	return out
}

func parseInstruction(line string) *Instruction {
	parts := strings.Split(line, " ")

	padding := 0
	if !IsOpcode(parts[0]) {
		// Skip label from parse
		padding = 1
	}

	i := &Instruction{
		Opcode: Opcode(parts[0+padding]),
	}

	if len(parts) > 1+padding {
		i.Op1 = parts[1+padding]
	}
	if len(parts) > 2+padding {
		i.Op2 = parts[2+padding]
	}
	if len(parts) > 3+padding {
		i.Op3 = parts[3+padding]
	}

	return i
}

// in Decoded instruction
func executeAddCalc(in chan *Instruction) chan *Instruction {
	s := stages[2]
	out := make(chan *Instruction, 5)
	go func() {
		Debug("Execute goroutine started and is waiting for messages\n")
		for instruction := range in {
			Debug("Execute Address Calculation recieved instruction %v\n", instruction)
			s.CurrInstruction = instruction
			s.IsActive = true

			switch instruction.Opcode {
			case HALT:
				fmt.Println("HALT")
				os.Exit(1)
			case ADDI:
				AddiOperation(instruction)
			case ADD:
				AddOperation(instruction)
			case BEQ:
				BeqOperation(instruction)
			}

			for {
				select {
				case <-s.UserChan:
					out <- instruction
					s.CurrInstruction = nil
					s.IsActive = false
				}
				break
			}
		}
		Debug("Execute will not recieve anything else\n")
		close(out)
	}()
	return out
}

// in Instruction after execution complete channel
func memoryAccess(in chan *Instruction) chan *Instruction {
	s := stages[3]
	out := make(chan *Instruction, 5)
	go func() {
		Debug("Memory Access goroutine started and is waiting for messages\n")
		for instruction := range in {
			Debug("Memory Access recieved instruction %v\n", instruction)
			s.CurrInstruction = instruction
			s.IsActive = true
			for {
				select {
				case <-s.UserChan:
					out <- instruction
					s.CurrInstruction = nil
					s.IsActive = false
				}
				break
			}
		}
		Debug("Memory Access will not recieve anything else\n")
		close(out)
	}()
	return out
}

// in Instruction after save
func writeBack(in chan *Instruction) chan *Instruction {
	stage := stages[4]
	// Arbitrary for now. This can cause issues for extensive jumping, filling this channel
	out := make(chan *Instruction, 256)
	go func() {
		Debug("Write Back goroutine started and is waiting for messages\n")
		for instruction := range in {
			Debug("Write Back recieved instruction %v\n", instruction)
			stage.CurrInstruction = instruction
			stage.IsActive = true
			for {
				select {
				case <-stage.UserChan:
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

func handleUserInput() {
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
				broadcast(char)
			case 'h', 'H':
				fmt.Println("Aperte V para ver os estágios, K para avançar o estágio, H para ajuda e Q para sair")
			}
		}
	}()
}

func main() {

	registers = make(map[string]int8)
	for i := 0; i <= numRegisters; i++ {
		nick := fmt.Sprintf("R%d", i)
		registers[nick] = 0
	}

	for i := 0; i < len(stages); i++ {
		stages[i] = &Stage{
			UserChan: make(chan rune),
			IsActive: false,
			CurrPC:   0,
		}
	}

	handleUserInput()

	pipeline = NewPipeline()
	instructionsChan := make(chan int, pipeline.Lines)

	decodeChan := instructionFetch(instructionsChan)
	executeChan := decodeInstruction(decodeChan)
	memAccessChan := executeAddCalc(executeChan)
	writeBackChan := memoryAccess(memAccessChan)
	out := writeBack(writeBackChan)

	for pipeline.PC != pipeline.Lines {
		pipeline.PC++
		instructionsChan <- pipeline.PC
		Debug("Send instruction from PC %d\n", pipeline.PC)
	}
	Debug("All instructions sended")
	close(instructionsChan)

	for o := range out {
		Debug("Instruction completed: %v\n", o)
	}

	fmt.Println("All instructions executed")
}
