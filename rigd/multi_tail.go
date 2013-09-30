package main

import (
	"github.com/gocardless/rig"
	"container/ring"
)

type TailIterator struct {
	start *ring.Ring
	cur   *ring.Ring
	done  bool
}

func NewTailIterator(r *ring.Ring) *TailIterator {
	return &TailIterator{start: r, cur: r, done: false}
}

func (it *TailIterator) peek() *rig.ProcessOutputMessage {
	if it.done || it.cur.Value == nil {
		return nil
	}
	msg := it.cur.Value.(rig.ProcessOutputMessage)
	return &msg
}

func (it *TailIterator) next() *rig.ProcessOutputMessage {
	if it.done || it.cur.Value == nil {
		return nil
	}
	if it.cur.Prev() == it.start {
		it.done = true
	}
	msg := it.cur.Value.(rig.ProcessOutputMessage)
	it.cur = it.cur.Prev()
	return &msg
}

func getNextMessage(iterators []*TailIterator) *rig.ProcessOutputMessage {
	nextIdx := 0
	for i, it := range iterators {
		msg, cur := it.peek(), iterators[nextIdx].peek()
		if msg == nil {
			continue
		}
		if cur == nil || msg.Time.After(cur.Time) {
			nextIdx = i
		}
	}
	return iterators[nextIdx].next()
}

func MultiTail(buffers []*ring.Ring, num int) []*rig.ProcessOutputMessage {
	tail := make([]*rig.ProcessOutputMessage, num)
	var iterators []*TailIterator
	for _, buf := range buffers {
		iterators = append(iterators, NewTailIterator(buf))
	}

	for i := num - 1; i >= 0; i -= 1 {
		msg := getNextMessage(iterators)
		if msg == nil {
			tail = tail[i + 1:num]
			break
		}
		tail[i] = msg
	}

	return tail
}

