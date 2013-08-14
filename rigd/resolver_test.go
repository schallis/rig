package main

import (
	"fmt"
	"github.com/gocardless/rig"
	"testing"
)

// === Simple resolution tests

func Test_ResolvingCompleteDescriptor(t *testing.T) {
	checkSimpleResolution(t, "stack1:service1:process1", &rig.Descriptor{
		Stack:   "stack1",
		Service: "service1",
		Process: "process1",
	})
}

func Test_ResolvingDescriptorWithInvalidProcess(t *testing.T) {
	res := NewResolver(buildConfig(), "stack1:service1:xprocess", "/")

	if _, err := res.GetDescriptor(); err == nil {
		t.Error("Expected a resolution error")
	}
}

func Test_ResolvingDescriptorWithoutProcess(t *testing.T) {
	checkSimpleResolution(t, "stack1:service1", &rig.Descriptor{
		Stack:   "stack1",
		Service: "service1",
	})
}

func Test_ResolvingDescriptorWithInvalidService(t *testing.T) {
	res := NewResolver(buildConfig(), "stack1:xservice", "/")

	if _, err := res.GetDescriptor(); err == nil {
		t.Error("Expected a resolution error")
	}
}

func Test_ResolvingDescriptorWithoutService(t *testing.T) {
	checkSimpleResolution(t, "stack1", &rig.Descriptor{
		Stack: "stack1",
	})
}

func Test_ResolvingDescriptorWithInvalidStack(t *testing.T) {
	res := NewResolver(buildConfig(), "xstack", "/")

	if _, err := res.GetDescriptor(); err == nil {
		t.Error("Expected a resolution error")
	}
}

func Test_ResolvingBlankDescriptor(t *testing.T) {
	res := NewResolver(buildConfig(), "", "/")

	if _, err := res.GetDescriptor(); err == nil {
		t.Error("Expected a resolution error")
	}
}

// === Default stack resolution tests

func Test_ResolvingDefaultService(t *testing.T) {
	checkSimpleResolution(t, "service3", &rig.Descriptor{
		Stack:   "default",
		Service: "service3",
	})
}

// === Contextual resolution tests

func checkSimpleResolution(t *testing.T, str string, example *rig.Descriptor) {
	res := NewResolver(buildConfig(), str, "/")

	d, err := res.GetDescriptor()
	if err != nil {
		t.Error("Unexpected resolution error")
		return
	}

	if *d != *example {
		t.Errorf("Expected %+v, got %+v", example, d)
	}
}

func checkContextualResolution(t *testing.T, str string, example *rig.Descriptor) {
	res := NewResolver(NewConfig(), str, "/")

	d, err := res.GetDescriptor()
	if err != nil {
		t.Error("Unexpected resolution error")
	}

	if *d != *example {
		t.Errorf("Expected %+v, got %+v", example, d)
	}
}

// === Config setup helpers

func buildConfig() *Config {
	config := NewConfig()

	stack := NewStack("stack1")
	config.Stacks[stack.Name] = stack
	addService(stack, 1)
	addService(stack, 2)

	defaultStack := NewStack("default")
	config.Stacks["default"] = defaultStack
	addService(defaultStack, 3)

	return config
}

func addService(stack *Stack, num int) {
	service := &Service{
		Name:      fmt.Sprintf("service%d", num),
		Dir:       fmt.Sprintf("/projects/srv%d", num),
		Stack:     stack,
		Processes: make(map[string]*Process),
	}
	stack.Services[service.Name] = service

	process := NewProcess("process1", "echo hello", service)
	service.Processes[process.Name] = process
}
