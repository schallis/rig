package main

import (
	"github.com/gocardless/rig/utils"
	"time"
)

type ProcessOutputMessage struct {
	Content string
	Stack   string
	Service string
	Process string
	Time    time.Time
}

type ProcessOutputSubscription struct {
	id         string
	dispatcher *ProcessOutputDispatcher
	msgCh      chan ProcessOutputMessage
	endCh      chan bool
}

func (s *ProcessOutputSubscription) End() {
	delete(s.dispatcher.subscriptions, s.id)
}

type ProcessOutputDispatcher struct {
	subscriptions map[string]*ProcessOutputSubscription
}

func NewProcessOutputDispatcher() *ProcessOutputDispatcher {
	return &ProcessOutputDispatcher{
		subscriptions: make(map[string]*ProcessOutputSubscription),
	}
}

func (d *ProcessOutputDispatcher) Subscribe(c chan ProcessOutputMessage) *ProcessOutputSubscription {
	s := &ProcessOutputSubscription{
		id:         utils.GenerateId(),
		dispatcher: d,
		msgCh:      c,
		endCh:      make(chan bool),
	}

	d.subscriptions[s.id] = s
	return s
}

func (d *ProcessOutputDispatcher) Publish(message ProcessOutputMessage) {
	for _, s := range d.subscriptions {
		s.msgCh <- message
	}
}

func (d *ProcessOutputDispatcher) End() {
	for _, s := range d.subscriptions {
		s.endCh <- true
	}
}
