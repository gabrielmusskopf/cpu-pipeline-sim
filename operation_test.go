package main

import "testing"

type PipelineNOOP struct {
	PC     int
	Labels map[string]int
}

func (p *PipelineNOOP) Read(pc int) string {
	return "two .fill 2"
}

func (p *PipelineNOOP) Label(label string) (int, bool) {
	pc, ok := p.Labels[label]
	return pc, ok
}

func (p *PipelineNOOP) JumpTo(pc int) {
	p.PC = pc
}

func (p *PipelineNOOP) Broadcast(r rune) {}

func (p *PipelineNOOP) Stages() []*Stage {
	return []*Stage{}
}

func TestAddi(t *testing.T) {
    var want int8 = 2

	pipeline := &PipelineNOOP{
		Labels: make(map[string]int),
	}

	registers = make(map[string]int8)
	registers["R0"] = 0
	registers["R1"] = 0
	registers["R2"] = 2

	instruction := &Instruction{
		Op1: "R0",
		Op2: "R1",
		Op3: "R2",
	}

	AddiOperation(instruction, pipeline)

	got := registers["R1"]
	if got != want {
		t.Errorf("ADDI = %d, want %d", got, want)
	}
}

func TestAddiLabeled(t *testing.T) {
    var want int8 = 2

	pipeline := &PipelineNOOP{
		Labels: map[string]int{
			"two": 10,
		},
	}

	registers = make(map[string]int8)
	registers["R0"] = 0
	registers["R1"] = 0

	instruction := &Instruction{
		Op1: "R0",
		Op2: "R1",
		Op3: "two",
	}

	AddiOperation(instruction, pipeline)

	got := registers["R1"]
	if got != want {
		t.Errorf("ADDI = %d, want %d", got, want)
	}
}

func TestAdd(t *testing.T) {
    var want int8 = 4

	pipeline := &PipelineNOOP{}

	registers = make(map[string]int8)
	registers["R1"] = 0
	registers["R2"] = 1
	registers["R3"] = 3

	instruction := &Instruction{
		Op1: "R1",
		Op2: "R2",
		Op3: "R3",
	}

	AddOperation(instruction, pipeline)

	got := registers["R1"]
	if got != want {
		t.Errorf("ADDI = %d, want %d", got, want)
	}
}

func TestBeq(t *testing.T) {
    var want int = 10

	labels := make(map[string]int)
	labels["loop"] = 10
	pipeline := &PipelineNOOP{
		PC:     0,
		Labels: labels,
	}

	registers = make(map[string]int8)
	registers["R1"] = 3
	registers["R2"] = 3

	instruction := &Instruction{
		Op1: "R1",
		Op2: "R2",
		Op3: "loop",
	}

	BeqOperation(instruction, pipeline)

	got := pipeline.PC
	if got != want {
		t.Errorf("BEQ jumped to %d, want %d", got, want)
	}
}

func TestSubi(t *testing.T) {
    var want int8 = 1

	pipeline := &PipelineNOOP{
		Labels: make(map[string]int),
	}

	registers = make(map[string]int8)
	registers["R9"] = 2
	registers["R10"] = 0
	registers["R11"] = 1

	instruction := &Instruction{
		Op1: "R9",
		Op2: "R10",
		Op3: "R11",
	}

	SubiOperation(instruction, pipeline)

	got := registers["R10"]
	if got != want {
		t.Errorf("SUBI = %d, want %d", got, want)
	}
}

func TestSubiLabeled(t *testing.T) {
    var want int8 = 2

	pipeline := &PipelineNOOP{
		Labels: map[string]int{
			"two": 10,
		},
	}

	registers = make(map[string]int8)
	registers["R9"] = 4
	registers["R10"] = 0

	instruction := &Instruction{
		Op1: "R9",
		Op2: "R10",
		Op3: "two",
	}

	SubiOperation(instruction, pipeline)

	got := registers["R10"]
	if got != want {
		t.Errorf("SUBI = %d, want %d", got, want)
	}
}
