package main

import (
	"fmt"
	"strconv"
	"strings"
)

func getRegisterName(r string) string {
	if strings.HasPrefix(r, "R") {
		return r
	}
	return fmt.Sprintf("R%s", r)
}

// "Jump" to PC and get the value
func spy(pc int, pipe Pipeline) int8 {
	line := pipe.Read(pc)
	parts := strings.Split(line, " ")
	r, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		fmt.Printf("ERROR: R3 is not a number. Maybe it was not decoded properly")
		return 0
	}
	return int8(r)
}

// Substiuindo lw: addi R0 R1 -1 = Soma R0 com neg1 e coloca no R1
func AddiOperation(i *Instruction, pipe Pipeline) error {
	op1, ok := registers[getRegisterName(i.Op1)]
	if !ok {
		i.Valid = false
		return fmt.Errorf("Register %s does not exist", i.Op1)
	}
	_, ok = registers[getRegisterName(i.Op2)]
	if !ok {
		i.Valid = false
		return fmt.Errorf("Register %s does not exist", i.Op2)
	}
	var op3 int8
	pc, ok := pipe.Label(i.Op3)
	if ok {
		// Contains a label. "Jump" to related PC to get the int8 value
		op3 = spy(pc, pipe)
	} else {
		op3 = registers[getRegisterName(i.Op3)]
	}
	registers[i.Op2] = op1 + op3
	return nil
}

// add R0 R1 R2
// R0 = R1 + R2
func AddOperation(i *Instruction, pip Pipeline) error {
	op1Nick := getRegisterName(i.Op1)
	op2Nick := getRegisterName(i.Op2)
	op3Nick := getRegisterName(i.Op3)

	_, ok := registers[op1Nick]
	if !ok {
		i.Valid = false
		return fmt.Errorf("ERROR: Register %s does not exist\n", i.Op1)
	}
	op2, ok := registers[op2Nick]
	if !ok {
		i.Valid = false
		return fmt.Errorf("ERROR: Register %s does not exist\n", i.Op1)
	}
	op3, ok := registers[op3Nick]
	if !ok {
		i.Valid = false
		return fmt.Errorf("ERROR: Register %s does not exist\n", i.Op1)
	}

	registers[op1Nick] = op2 + op3
	return nil
}

func SubiOperation(i *Instruction, pipe Pipeline) error {
	op1, ok := registers[getRegisterName(i.Op1)]
	if !ok {
		i.Valid = false
		return fmt.Errorf("Register %s does not exist", i.Op1)
	}
	_, ok = registers[getRegisterName(i.Op2)]
	if !ok {
		i.Valid = false
		return fmt.Errorf("Register %s does not exist", i.Op2)
	}
	var op3 int8
	pc, ok := pipe.Label(i.Op3)
	if ok {
		// Contains a label. "Jump" to related PC to get the int8 value
		op3 = spy(pc, pipe)
	} else {
		op3 = registers[getRegisterName(i.Op3)]
	}
	registers[i.Op2] = op1 - op3
	return nil
}

// sub R0 R1 R2
// R0 = R1 - R2
func SubOperation(i *Instruction, pip Pipeline) error {
	op1Nick := getRegisterName(i.Op1)
	op2Nick := getRegisterName(i.Op2)
	op3Nick := getRegisterName(i.Op3)

	_, ok := registers[op1Nick]
	if !ok {
		i.Valid = false
		return fmt.Errorf("ERROR: Register %s does not exist\n", i.Op1)
	}
	op2, ok := registers[op2Nick]
	if !ok {
		i.Valid = false
		return fmt.Errorf("ERROR: Register %s does not exist\n", i.Op1)
	}
	op3, ok := registers[op3Nick]
	if !ok {
		i.Valid = false
		return fmt.Errorf("ERROR: Register %s does not exist\n", i.Op1)
	}

	registers[op1Nick] = op2 - op3
	return nil
}

func BeqOperation(i *Instruction, pipe Pipeline) error {
	op1Nick := getRegisterName(i.Op1)
	op2Nick := getRegisterName(i.Op2)

	op1, ok := registers[op1Nick]
	if !ok {
		i.Valid = false
		return fmt.Errorf("ERROR: Register %s does not exist\n", i.Op1)
	}
	op2, ok := registers[op2Nick]
	if !ok {
		i.Valid = false
		return fmt.Errorf("ERROR: Register %s does not exist\n", i.Op1)
	}
	if op1 == op2 {
		pc, ok := pipe.Label(i.Op3)
		if !ok {
			return fmt.Errorf("ERROR: Label %s does not exist\n", i.Op3)
		}
		Debug("Jumping to %d\n", pc)
		pipe.JumpTo(pc)
	}
	return nil
}

// Jump to labeled PC
// j loop
func JOperation(i *Instruction, pipe Pipeline) error {
	pc, ok := pipe.Label(i.Op1)
	if !ok {
		return fmt.Errorf("ERROR: Label %s does not exist\n", i.Op3)
	}

	Debug("Jumping to %d\n", pc)
	pipe.JumpTo(pc)
	return nil
}
