package logging

import (
	"time"
	"fmt"
)

type Message struct {
	Content interface{}
	Source  string
	Time    time.Time
}

type Dispatcher struct {
	sink        chan Message
	subscribers []chan<- Message
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{sink: make(chan Message)}
}

func (d *Dispatcher) Start() {
	for msg := range(d.sink) {
		for _, subscriber := range(d.subscribers) {
			go func(subscriber chan<- Message, msg Message) {
				subscriber <- msg
			}(subscriber, msg)
		}
	}
}

func (d *Dispatcher) Stop() {
	close(d.sink)
}

func (d *Dispatcher) Register(ch chan<- Message) {
	d.subscribers = append(d.subscribers, ch)
}

type Logger struct {
	sink   chan Message
	source string
}

func NewLogger(d *Dispatcher, source string) *Logger {
	return &Logger{sink: d.sink, source: source}
}

func (l *Logger) Log(content interface{}) {
	l.sink <- Message{Content: content, Source: l.source, Time: time.Now()}
}

func (l *Logger) Logf(format string, v ...interface{}) {
	content := fmt.Sprintf(format, v...)
	l.sink <- Message{Content: content, Source: l.source, Time: time.Now()}
}

