package main

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
