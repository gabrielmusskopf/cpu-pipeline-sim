package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

var debug = false

func Debug(format string, v ...any) {
	if debug {
		log.Printf(format, v...)
	}
}

type Pipeline interface {
	Read(int) string
	Label(string) (int, bool)
	JumpTo(int)
	Broadcast(rune)
	Stages() []*Stage
}

type PipelineFile struct {
	File   []byte
	PC     int
	Labels map[string]int // Label: PC
	Lines  int
	In     chan int
	Out    chan *Instruction
	s      []*Stage
}

func NewPipeline() *PipelineFile {
	file, err := os.Open("instrucoes.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	b, err := io.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}
	pipeline := &PipelineFile{
		File: b,
		PC:   0,
		In:   make(chan int),
	}

	pipeline.ParseFile()

	pipeline.s = []*Stage{
		NewStage("Instruction fetch", "fet"),
		NewStage("Decode instruction", "dec"),
		NewStage("Execute instruction", "exe"),
		NewStage("Memory access", "mem"),
		NewStage("Write back", "wrb"),
	}

	decodeChan := pipeline.instructionFetch(pipeline.In)
	executeChan := pipeline.decodeInstruction(decodeChan)
	memAccessChan := pipeline.executeAddCalc(executeChan)
	writeBackChan := pipeline.memoryAccess(memAccessChan)
	pipeline.Out = pipeline.writeBack(writeBackChan)

	return pipeline
}

func (p *PipelineFile) ParseFile() {
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

func (p *PipelineFile) Read(num int) string {
	reader := bytes.NewReader(p.File)
	sc := bufio.NewScanner(reader)
	for i := 1; i <= num; i++ {
		sc.Scan()
	}
	return sc.Text()
}

func (p *PipelineFile) Label(name string) (int, bool) {
	pc, ok := p.Labels[name]
	return pc, ok
}

func (p *PipelineFile) JumpTo(pc int) {
	p.PC = pc
}

func (p *PipelineFile) Broadcast(v rune) {
	for _, stage := range p.s {
		if stage.IsActive {
			stage.UserChan <- v
		}
	}
}

func (p PipelineFile) Stages() []*Stage {
	return p.s
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
	HALT Opcode = "halt"
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

type Stage struct {
	Name            string
	Nickname        string
	UserChan        chan rune
	CurrInstruction *Instruction
	CurrPC          int
	IsActive        bool
}

func NewStage(name, nc string) *Stage {
	return &Stage{
		Name:     name,
		Nickname: nc,
		UserChan: make(chan rune),
		IsActive: false,
		CurrPC:   0,
	}
}

// in Program counter (PC)
func (p *PipelineFile) instructionFetch(in chan int) chan string {
    s := p.s[0]
	out := make(chan string)
	go func() {
		Debug("%s goroutine started and is waiting for messages\n", s.Name)
		for pc := range in {
			Debug("Instruction fetch recieved PC %d\n", pc)
			s.CurrPC = pc
			s.IsActive = true
			instruction := p.Read(pc)
			for {
				select {
				case <-s.UserChan:
					out <- instruction
					s.IsActive = false
				}
				break
			}
		}
		Debug("%s will not recieve anything else\n", s.Name)
		close(out)
	}()

	return out
}

// in Raw instrucion line channel
func (p *PipelineFile) decodeInstruction(in chan string) chan *Instruction {
    s := p.s[1]
	out := make(chan *Instruction)
	go func() {
		Debug("%s goroutine started and is waiting for messages\n", s.Name)
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
		Debug("%s will not recieve anything else\n", s.Name)
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
func (p *PipelineFile) executeAddCalc(in chan *Instruction) chan *Instruction {
    s := p.s[2]
	out := make(chan *Instruction)
	go func() {
		Debug("%s goroutine started and is waiting for messages\n", s.Name)
		for instruction := range in {
			Debug("Execute Address Calculation recieved instruction %v\n", instruction)
			s.CurrInstruction = instruction
			s.IsActive = true

			switch instruction.Opcode {
			case HALT:
				fmt.Println("HALT")
				os.Exit(1)
			case ADDI:
				AddiOperation(instruction, p)
			case ADD:
				AddOperation(instruction, p)
			case BEQ:
				BeqOperation(instruction, p)
			case SUB:
				SubiOperation(instruction, p)
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
		Debug("%s will not recieve anything else\n", s.Name)
		close(out)
	}()
	return out
}

// in Instruction after execution complete channel
func (p *PipelineFile) memoryAccess(in chan *Instruction) chan *Instruction {
    s := p.s[3]
	out := make(chan *Instruction)
	go func() {
		Debug("%s goroutine started and is waiting for messages\n", s.Name)
		for instruction := range in {
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
		Debug("%s will not recieve anything else\n", s.Name)
		close(out)
	}()
	return out
}

// in Instruction after save
func (p *PipelineFile) writeBack(in chan *Instruction) chan *Instruction {
    s := p.s[4]
	out := make(chan *Instruction)
	go func() {
		Debug("%s goroutine started and is waiting for messages\n", s.Name)
		for instruction := range in {
			Debug("Write Back recieved instruction %v\n", instruction)
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
		Debug("%s will not recieve anything else\n", s.Name)
		close(out)
	}()
	return out
}

func main() {
	registers = make(map[string]int8)
	for i := 0; i <= numRegisters; i++ {
		nick := fmt.Sprintf("R%d", i)
		registers[nick] = 0
	}

	pipeline := NewPipeline()

	term := &Terminal{
		Pipeline: pipeline,
	}
	term.HandleUserInput()

	go func() {
		for o := range pipeline.Out {
			Debug("Instruction completed: %v\n", o)
		}
	}()

	for pipeline.PC != pipeline.Lines {
		pipeline.PC++
		pipeline.In <- pipeline.PC
		Debug("Send instruction from PC %d\n", pipeline.PC)
	}
	Debug("All instructions sended")
	close(pipeline.In)

	fmt.Println("All instructions executed")
}
