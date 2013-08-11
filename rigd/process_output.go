package main

import (
	"crypto/rand"
	"encoding/hex"
	"io"
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
		id:         generateId(),
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

// https://github.com/dotcloud/docker/blob/940d58806c3e3d4409a7eee4859335e98139d09f/image.go#L218-225
func generateId() string {
	id := make([]byte, 32)
	_, err := io.ReadFull(rand.Reader, id)
	if err != nil {
		panic(err) // This shouldn't happen
	}
	return hex.EncodeToString(id)
}
