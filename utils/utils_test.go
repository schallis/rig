package utils

import (
	"os/user"
	"testing"
)

type expandPathTest struct {
	initialPath  string
	expandedPath string
}

func TestExpandPath(t *testing.T) {
	currentUser, err := user.Current()
	if err != nil {
		t.Errorf("Unknown user : %v", err)
	}
	username := currentUser.Username
	homeDir := currentUser.HomeDir

	expandPathTests := []*expandPathTest{
		&expandPathTest{"~/test", homeDir + "/test"},
		&expandPathTest{"~" + username + "/test", homeDir + "/test"},
	}

	for i, test := range expandPathTests {
		t.Logf("Expanding : %s", test.initialPath)
		expandedPath := ExpandPath(test.initialPath)
		if expandedPath != test.expandedPath {
			t.Errorf("Test %d : %s should equal %s", i, expandedPath, test.expandedPath)
		}
	}
}
