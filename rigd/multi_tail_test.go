package main

import (
	"github.com/gocardless/rig"
	"time"
	"testing"
	"container/ring"
)

func Test_SingleTail(t *testing.T) {
	buf := ring.New(3)
	buf.Value = rig.ProcessOutputMessage{Content: "a", Time: time.Now()}
	buf = buf.Next()
	buf.Value = rig.ProcessOutputMessage{Content: "b", Time: time.Now().Add(1)}
	buf = buf.Next()
	buf.Value = rig.ProcessOutputMessage{Content: "c", Time: time.Now().Add(2)}

	tail := MultiTail([]*ring.Ring{buf}, 2)
	if len(tail) != 2 {
		t.Errorf("Expected len(tail) to be 2, got %d", len(tail))
	}
	if tail[0].Content != "b" {
		t.Errorf("Expected tail[0] to be 'b', got '%v'", tail[0].Content)
	}
	if tail[1].Content != "c" {
		t.Errorf("Expected tail[1] to be 'c', got '%v'", tail[1].Content)
	}
}

func Test_MultiTail(t *testing.T) {
	buf1 := ring.New(2)
	buf1.Value = rig.ProcessOutputMessage{Content: "a", Time: time.Now()}

	buf2 := ring.New(2)
	buf2.Value = rig.ProcessOutputMessage{Content: "b", Time: time.Now().Add(1)}

	buf1 = buf1.Next()
	buf1.Value = rig.ProcessOutputMessage{Content: "c", Time: time.Now().Add(2)}

	buf2 = buf2.Next()
	buf2.Value = rig.ProcessOutputMessage{Content: "d", Time: time.Now().Add(3)}

	tail := MultiTail([]*ring.Ring{buf1, buf2}, 4)
	if len(tail) != 4 {
		t.Errorf("Expected len(tail) to be 2, got %d", len(tail))
	}
	if tail[0].Content != "a" {
		t.Errorf("Expected tail[0] to be 'a', got '%v'", tail[0].Content)
	}
	if tail[1].Content != "b" {
		t.Errorf("Expected tail[1] to be 'b', got '%v'", tail[1].Content)
	}
	if tail[2].Content != "c" {
		t.Errorf("Expected tail[2] to be 'c', got '%v'", tail[2].Content)
	}
	if tail[3].Content != "d" {
		t.Errorf("Expected tail[3] to be 'c', got '%v'", tail[3].Content)
	}
}
