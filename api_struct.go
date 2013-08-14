package rig

import (
	"time"
)

type ApiVersion struct {
	Version string
}

type ApiProcess struct {
	Name   string
	Pid    int
	Status int
}

type Descriptor struct {
	Stack   string
	Service string
	Process string
}

type ProcessOutputMessage struct {
	Content string
	Stack   string
	Service string
	Process string
	Time    time.Time
}
