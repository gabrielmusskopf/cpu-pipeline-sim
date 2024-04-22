package main

import "strings"

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
