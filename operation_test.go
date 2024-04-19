package main

import "testing"

type PipelineNOOP struct {
	PC     int
	Labels map[string]int
}

func (p *PipelineNOOP) Read(pc int) string {
	return "neg1 .fill 2"
}

func (p *PipelineNOOP) Label(label string) (int, bool) {
	pc, ok := p.Labels[label]
	return pc, ok
}

func (p *PipelineNOOP) CurrPC() int {
    return p.PC
}

func (p *PipelineNOOP) Next() int {
	return 0
}

func (p *PipelineNOOP) Close() {}

func (p *PipelineNOOP) Pull() chan *Instruction {
	return nil
}

func (p *PipelineNOOP) IsOver() bool {
	return true
}

func (p *PipelineNOOP) JumpTo(pc int) {
    p.PC = pc
}

func TestAddi(t *testing.T) {
	pipeline = &PipelineNOOP{
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

	AddiOperation(instruction)

	got := registers["R1"]
	if got != 2 {
		t.Errorf("ADDI = %d, want 2", got)
	}
}

func TestAddiLabeled(t *testing.T) {
	labels := make(map[string]int)
	labels["two"] = 10
	pipeline = &PipelineNOOP{
		Labels: labels,
	}

	registers = make(map[string]int8)
	registers["R0"] = 0
	registers["R1"] = 0

	instruction := &Instruction{
		Op1: "R0",
		Op2: "R1",
		Op3: "two",
	}

	AddiOperation(instruction)

	got := registers["R1"]
	if got != 2 {
		t.Errorf("ADDI = %d, want 2", got)
	}
}

func TestAdd(t *testing.T) {
	pipeline = &PipelineNOOP{}

	registers = make(map[string]int8)
	registers["R0"] = 0
	registers["R1"] = 0
	registers["R2"] = 3

	instruction := &Instruction{
		Op1: "R0",
		Op2: "R1",
		Op3: "R2",
	}

	AddiOperation(instruction)

	got := registers["R1"]
	if got != 3 {
		t.Errorf("ADDI = %d, want 3", got)
	}
}

func TestBeq(t *testing.T) {
	labels := make(map[string]int)
	labels["loop"] = 10
	pipeline = &PipelineNOOP{
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

	BeqOperation(instruction)

	got := pipeline.CurrPC()
	if got != 10 {
		t.Errorf("BEQ jumped to %d, want 10", got)
	}
}
