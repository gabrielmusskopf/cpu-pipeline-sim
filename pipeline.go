package main

import (
	"log"
	"os"
	"strings"
)

type Pipeline interface {
	Read(int) string
	Label(string) (int, bool)
	JumpTo(int)
	Broadcast(rune)
	Stages() []*Stage
}

type PipelineFile struct {
	Lines  []string
	PC     int
	Labels map[string]int // Label: PC
	In     chan int
	Out    chan *Instruction
	s      []*Stage
}

func NewPipeline(filename string) *PipelineFile {
	b, err := os.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}

	// For some reason, bytes to string cast cause an extra '\n'
	content, _ := strings.CutSuffix(string(b), "\n")
	lines := strings.Split(content, "\n")

	pipeline := &PipelineFile{
		Lines: lines,
		PC:    0,
		In:    make(chan int),
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
	labels := make(map[string]int)

	for i, line := range p.Lines {
		lineParts := strings.Split(line, " ")
		key := lineParts[0]

		if !IsOpcode(key) {
			labels[key] = i + 1
			Debug("Parsed [%s: %d] constant\n", key, labels[key])
		}
	}

	p.Labels = labels
}

func (p *PipelineFile) Start() {
	go func() {
		for o := range p.Out {
			Info("Instruction completed: %v\n", o)
		}
	}()

	go func() {
		for p.PC != len(p.Lines) {
			p.PC++
			p.In <- p.PC
			Debug("Send instruction from PC %d\n", p.PC)
		}
		Info("All instructions sended")
		close(p.In)

		Info("All instructions executed")
	}()
}

func (p *PipelineFile) Read(num int) string {
	if num > len(p.Lines) {
		return ""
	}
	return p.Lines[num-1]
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

			events <- stageToggledMsg{
				position: 0,
				value:    pc,
			}

			<-s.UserChan
			out <- instruction
			s.IsActive = false
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

			events <- stageToggledMsg{
				position: 1,
				value:    instruction,
			}

			<-s.UserChan
			out <- instruction
			s.CurrInstruction = nil
			s.IsActive = false
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
				Debug("HALT!\n")
				events <- quitMsg{}
			case ADDI:
				AddiOperation(instruction, p)
			case ADD:
				AddOperation(instruction, p)
			case BEQ:
				BeqOperation(instruction, p)
			case SUBI:
				SubiOperation(instruction, p)
			case SUB:
				SubOperation(instruction, p)
			case J:
				JOperation(instruction, p)
			}

			events <- stageToggledMsg{
				position: 2,
				value:    instruction,
			}

			<-s.UserChan
			out <- instruction
			s.CurrInstruction = nil
			s.IsActive = false
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

			events <- stageToggledMsg{
				position: 3,
				value:    instruction,
			}

			<-s.UserChan
			out <- instruction
			s.CurrInstruction = nil
			s.IsActive = false
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

			events <- stageToggledMsg{
				position: 4,
				value:    instruction,
			}

			<-s.UserChan
			out <- instruction
			s.CurrInstruction = nil
			s.IsActive = false
		}
		Debug("%s will not recieve anything else\n", s.Name)
		close(out)
	}()
	return out
}
