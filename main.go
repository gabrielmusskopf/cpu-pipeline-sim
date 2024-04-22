package main

import (
	"fmt"
	"time"
)

var events chan interface{} = make(chan interface{}, 20)
var debug = true

func Debug(format string, v ...any) {
	if debug {
		message := fmt.Sprintf(format, v...)
        events <- debugMsg{message: fmt.Sprintf("%s %s", time.Now().Format("15:04:05 2006-01-02"), message)}
	}
}

var numRegisters = 32
var registers map[string]int8

func updateRegister(name string, value int8) {
	_, ok := registers[name]
	if !ok {
		return
	}
	registers[name] = value
	events <- registerUpdatedMsg{name: name, value: value}
}


func main() {
	registers = make(map[string]int8)
	for i := 0; i < numRegisters; i++ {
		nick := fmt.Sprintf("R%d", i)
		registers[nick] = 0
	}

	pipeline := NewPipeline()
    pipeline.Start()

	RunCmd(pipeline, registers, events)
}
