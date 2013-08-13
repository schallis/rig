package main

import (
	"github.com/gocardless/rig"
	"github.com/gocardless/rig/utils"
	"sync"
)

type ProcessOutputSubscription struct {
	id         string
	dispatcher *ProcessOutputDispatcher
	msgCh      chan rig.ProcessOutputMessage
	endCh      chan bool
}

func (s *ProcessOutputSubscription) End() {
	s.dispatcher.Lock()
	delete(s.dispatcher.subscriptions, s.id)
	s.dispatcher.Unlock()
}

type ProcessOutputDispatcher struct {
	sync.RWMutex
	subscriptions map[string]*ProcessOutputSubscription
}

func NewProcessOutputDispatcher() *ProcessOutputDispatcher {
	return &ProcessOutputDispatcher{
		subscriptions: make(map[string]*ProcessOutputSubscription),
	}
}

func (d *ProcessOutputDispatcher) Subscribe(c chan rig.ProcessOutputMessage) *ProcessOutputSubscription {
	s := &ProcessOutputSubscription{
		id:         utils.GenerateId(),
		dispatcher: d,
		msgCh:      c,
		endCh:      make(chan bool),
	}

	d.Lock()
	d.subscriptions[s.id] = s
	d.Unlock()
	return s
}

func (d *ProcessOutputDispatcher) Publish(message rig.ProcessOutputMessage) {
	d.RLock()
	for _, s := range d.subscriptions {
		s.msgCh <- message
	}
	d.RUnlock()
}

func (d *ProcessOutputDispatcher) End() {
	d.RLock()
	for _, s := range d.subscriptions {
		s.endCh <- true
	}
	d.RUnlock()
}
