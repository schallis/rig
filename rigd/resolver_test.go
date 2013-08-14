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
	withConfig(t, func(c *Config) {
		res := NewResolver(c, "stack1:service1:xprocess", "/")

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
	withConfig(t, func(c *Config) {
		res := NewResolver(c, "stack1:xservice", "/")

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
	withConfig(t, func(c *Config) {
		res := NewResolver(c, "xstack", "/")

		if _, err := res.GetDescriptor(); err == nil {
			t.Error("Expected a resolution error")
		}
	})
}

func Test_ResolvingBlankDescriptor(t *testing.T) {
	withConfig(t, func(c *Config) {
		res := NewResolver(c, "", "/")

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
	withConfig(t, func(c *Config) {
		dir := path.Join(c.Dir, "..", "projects", "srv2")
		res := NewResolver(c, ":xservice", dir)

		if _, err := res.GetDescriptor(); err == nil {
			t.Error("Expected a resolution error")
		}
	})
}


// === Test helpers

func checkSimpleResolution(t *testing.T, str string, example *rig.Descriptor) {
	withConfig(t, func(c *Config) {
		res := NewResolver(c, str, "/")

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
	withConfig(t, func(c *Config) {
		dir = path.Join(c.Dir, "..", dir)
		res := NewResolver(c, str, dir)

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

type testFunc func(*Config)

func withConfig(t *testing.T, f testFunc) *Config {
	tmpDir, err := ioutil.TempDir("", "resolver-test")
	if err != nil {
		t.Fatal("creating temp dir:", err)
	}
	//defer os.RemoveAll(tmpDir)

	projectDir := path.Join(tmpDir, "projects")
	os.Mkdir(projectDir, 0755)
	setupProjects(projectDir)

	configDir := path.Join(tmpDir, "config")
	os.Mkdir(configDir, 0755)

	os.Mkdir(path.Join(configDir, "stack1"), 0755)
	os.Symlink(
		path.Join(projectDir, "srv1"),
		path.Join(configDir, "stack1", "service1"),
	)
	os.Symlink(
		path.Join(projectDir, "srv2"),
		path.Join(configDir, "stack1", "service2"),
	)

	os.Symlink(path.Join(projectDir, "srv3"), path.Join(configDir, "service3"))

	config := NewConfigFromDir(configDir)

	os.Stdout.Sync()
	f(config)

	return config
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
