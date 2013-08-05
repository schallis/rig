package logging

import (
	"fmt"
	"strings"
)

const (
	PINK = "\x1b[35m"
	BLUE = "\x1b[34m"
	CYAN = "\x1b[36m"
	ORANGE = "\x1b[33m"
	GREEN = "\x1b[32m"
	RESET = "\x1b[0m"
)
var palette = []string{PINK, BLUE, CYAN, ORANGE, GREEN}

type TerminalSubscriber struct {
	nextColorIndex int
	colors         map[string]string
	sourceWidth    int
}

func NewTerminalSubscriber(d *Dispatcher, sourceWidth int) *TerminalSubscriber {
	ts := &TerminalSubscriber{
		nextColorIndex: 0,
		colors:         make(map[string]string),
		sourceWidth:    sourceWidth,
	}

	ts.start(d)
	return ts
}

func (ts *TerminalSubscriber) start(d *Dispatcher) {
	ch := make(chan Message)
	d.Register(ch)

	for msg := range(ch) {
		timeStr := msg.Time.Format("15:04:05")
		fmt.Print(ts.chooseColor(msg.Source))
		source := ts.padOrTrimSource(msg.Source)
		fmt.Printf("%v %v | ", timeStr, source)
		fmt.Print(RESET)
		fmt.Println(msg.Content)
	}
}

func (ts *TerminalSubscriber) padOrTrimSource(name string) string {
	diff := ts.sourceWidth - len(name)
	if diff <= 0 {
		name = name[0:ts.sourceWidth]
	} else {
		name += strings.Repeat(" ", diff)
	}
	return name
}

func (ts *TerminalSubscriber) chooseColor(source string) string {
	if color, ok := ts.colors[source]; ok {
		return color
	}
	color := palette[ts.nextColorIndex]
	ts.colors[source] = color

	ts.nextColorIndex++
	if ts.nextColorIndex >= len(palette) {
		ts.nextColorIndex = 0
	}

	return color
}

