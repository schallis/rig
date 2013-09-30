package main

import (
	"fmt"
	"github.com/gocardless/rig"
	"math"
)

var (
	colors = []string{
		"\x1b[36m", // cyan
		"\x1b[33m", // orange
		"\x1b[32m", // green
		"\x1b[35m", // pink
		"\x1b[34m", // blue
		"\x1b[37m", // white
	}
	errorColor string = "\x1b[31m"

	reset string = "\x1b[0m"
	bold  string = "\x1b[1m"
)

type ProcessLogger struct {
	processColor map[string]string
	colorCounter int
	maxMetaSize  float64
}

func NewProcessLogger() *ProcessLogger {
	return &ProcessLogger{
		processColor: make(map[string]string),
		colorCounter: 0,
		maxMetaSize:  0,
	}
}

func (p *ProcessLogger) Println(m rig.ProcessOutputMessage) {
	d := m.Stack + ":" + m.Service + ":" + m.Process
	if p.processColor[d] == "" {
		p.processColor[d] = colors[p.colorCounter]
		p.colorCounter++
		if p.colorCounter == len(colors) {
			p.colorCounter = 0
		}
	}
	color := p.processColor[d]

	str := fmt.Sprintf("%s", color)
	str += fmt.Sprintf("%s ", m.Time.Format("15:04:05"))

	meta := fmt.Sprintf("%s:%s:%s", m.Stack, m.Service, m.Process)
	p.maxMetaSize = math.Max(p.maxMetaSize, float64(len(meta)))
	str += meta

	str += fmt.Sprintf("%s | ", toSpace(p.maxMetaSize, meta))
	str += fmt.Sprintf("%s", m.Content)
	str += fmt.Sprintf("%s", reset)

	fmt.Println(str)
}

func toSpace(max float64, meta string) (spaces string) {
	diff := max - float64(len(meta))
	if diff > 0 {
		for i := 0; float64(i) < diff; i++ {
			spaces += " "
		}
	}
	return spaces
}
