package main

import (
	"github.com/gocardless/rig"
	"io/ioutil"
	"os"
	"path"
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
	withStacks(t, func(s map[string]*Stack) {
		res := NewResolver(s, "stack1:service1:xprocess", "/")

		if _, err := res.GetDescriptor(); err == nil {
			t.Error("Expected a resolution error")
		}
	})
}

func Test_ResolvingDescriptorWithoutProcess(t *testing.T) {
	checkSimpleResolution(t, "stack1:service1", &rig.Descriptor{
		Stack:   "stack1",
		Service: "service1",
	})
}

func Test_ResolvingDescriptorWithInvalidService(t *testing.T) {
	withStacks(t, func(s map[string]*Stack) {
		res := NewResolver(s, "stack1:xservice", "/")

		if _, err := res.GetDescriptor(); err == nil {
			t.Error("Expected a resolution error")
		}
	})
}

func Test_ResolvingDescriptorWithoutService(t *testing.T) {
	checkSimpleResolution(t, "stack1", &rig.Descriptor{
		Stack: "stack1",
	})
}

func Test_ResolvingDescriptorWithInvalidStack(t *testing.T) {
	withStacks(t, func(s map[string]*Stack) {
		res := NewResolver(s, "xstack", "/")

		if _, err := res.GetDescriptor(); err == nil {
			t.Error("Expected a resolution error")
		}
	})
}

func Test_ResolvingBlankDescriptor(t *testing.T) {
	withStacks(t, func(s map[string]*Stack) {
		res := NewResolver(s, "", "/")

		if _, err := res.GetDescriptor(); err == nil {
			t.Error("Expected a resolution error")
		}
	})
}

// === Default stack resolution tests

func Test_ResolvingDefaultService(t *testing.T) {
	checkSimpleResolution(t, "service3", &rig.Descriptor{
		Stack:   "default",
		Service: "service3",
	})
}

func Test_ResolvingDefaultServiceWithProcess(t *testing.T) {
	checkSimpleResolution(t, "service3:process1", &rig.Descriptor{
		Stack:   "default",
		Service: "service3",
		Process: "process1",
	})
}

// === Contextual resolution tests

func Test_ResolvingContextualService(t *testing.T) {
	checkContextualResolution(t, "", "projects/srv2", &rig.Descriptor{
		Stack:   "stack1",
		Service: "service2",
	})
}

func Test_ResolvingContextualServiceWithProcess(t *testing.T) {
	checkContextualResolution(t, "process1", "projects/srv2", &rig.Descriptor{
		Stack:   "stack1",
		Service: "service2",
		Process: "process1",
	})
}

// === Sibling resolution tests

func Test_ResolvingContextualSibling(t *testing.T) {
	checkContextualResolution(t, ":service1", "projects/srv2", &rig.Descriptor{
		Stack:   "stack1",
		Service: "service1",
	})
}

func Test_ResolvingContextualSiblingWithProcess(t *testing.T) {
	checkContextualResolution(t, ":service1:process1", "projects/srv2", &rig.Descriptor{
		Stack:   "stack1",
		Service: "service1",
		Process: "process1",
	})
}

func Test_ResolvingInvalidContextualSibling(t *testing.T) {
	withStacks(t, func(s map[string]*Stack) {
		dir := path.Join(tmpDir, "projects", "srv2")
		res := NewResolver(s, ":xservice", dir)

		if _, err := res.GetDescriptor(); err == nil {
			t.Error("Expected a resolution error")
		}
	})
}

// === Test helpers

func checkSimpleResolution(t *testing.T, str string, example *rig.Descriptor) {
	withStacks(t, func(s map[string]*Stack) {
		res := NewResolver(s, str, "/")

		d, err := res.GetDescriptor()
		if err != nil {
			t.Errorf("Resolution error: %v", err)
			return
		}

		if *d != *example {
			t.Errorf("Expected %+v, got %+v", example, d)
		}
	})
}

func checkContextualResolution(t *testing.T, str, dir string, example *rig.Descriptor) {
	withStacks(t, func(s map[string]*Stack) {
		dir = path.Join(tmpDir, dir)
		res := NewResolver(s, str, dir)

		d, err := res.GetDescriptor()
		if err != nil {
			t.Errorf("Resolution error: %v", err)
			return
		}

		if *d != *example {
			t.Errorf("Expected %+v, got %+v", example, d)
		}
	})
}

// === Config setup helpers

// type testFunc func(*Config)
type testFunc func(map[string]*Stack)

var tmpDir string

func withStacks(t *testing.T, f testFunc) {
	var err error
	tmpDir, err = ioutil.TempDir("", "resolver-test")
	if err != nil {
		t.Fatal("creating temp dir:", err)
	}
	defer os.RemoveAll(tmpDir)

	projectDir := path.Join(tmpDir, "projects")
	os.Mkdir(projectDir, 0755)
	setupProjects(projectDir)

	stack1 := NewStack("stack1")
	service1, _ := NewService("service1", path.Join(tmpDir, "projects", "srv1"), stack1)
	service2, _ := NewService("service2", path.Join(tmpDir, "projects", "srv2"), stack1)
	stack1.Services[service1.Name] = service1
	stack1.Services[service2.Name] = service2

	defaultStack := NewStack("default")
	service3, _ := NewService("service3", path.Join(tmpDir, "projects", "srv3"), defaultStack)
	defaultStack.Services[service3.Name] = service3

	stacks := map[string]*Stack{
		"default": defaultStack,
		"stack1":  stack1,
	}

	f(stacks)
}

func setupProjects(dir string) {
	procfile := []byte("process1: echo hello\n")

	os.Mkdir(path.Join(dir, "srv1"), 0755)
	ioutil.WriteFile(path.Join(dir, "srv1", "Procfile"), procfile, 0644)

	os.Mkdir(path.Join(dir, "srv2"), 0755)
	ioutil.WriteFile(path.Join(dir, "srv2", "Procfile"), procfile, 0644)

	os.Mkdir(path.Join(dir, "srv3"), 0755)
	ioutil.WriteFile(path.Join(dir, "srv3", "Procfile"), procfile, 0644)
}
