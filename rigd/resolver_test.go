package main

import (
	"github.com/gocardless/rig"
	"testing"
)

func checkSimpleResolution(t *testing.T, str string, example *rig.Descriptor) {
	res := NewDescriptorResolver(str, "/")

	d, err := res.GetDescriptor()
	if err != nil {
		t.Error("Unexpected resolution error")
	}

	if *d != *example {
		t.Errorf("Expected %+v, got %+v", example, d)
	}
}

func Test_ResolvingCompleteDescriptor(t *testing.T) {
	checkSimpleResolution(t, "stack1:service1:process1", &rig.Descriptor{
		Stack:   "stack1",
		Service: "service1",
		Process: "process1",
	})
}

func Test_ResolvingDescriptorWithoutProcess(t *testing.T) {
	checkSimpleResolution(t, "stack1:service1", &rig.Descriptor{
		Stack:   "stack1",
		Service: "service1",
	})
}

func Test_ResolvingDescriptorWithoutService(t *testing.T) {
	checkSimpleResolution(t, "stack1", &rig.Descriptor{
		Stack:   "stack1",
	})
}

func Test_ResolvingBlankDescriptor(t *testing.T) {
	res := NewDescriptorResolver("", "/")

	if _, err := res.GetDescriptor(); err == nil {
		t.Error("Expected a resolution error")
	}
}
