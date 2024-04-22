package main

type responseMsg struct{}

type quitMsg struct{
cause string
}

type stageToggledMsg struct {
	position int
	value    any
}

type registerUpdatedMsg struct {
	name  string
	value int8
}

type debugMsg struct {
	message string
}
