package utils

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"os/user"
	"regexp"
	"strings"
)

// https://github.com/dotcloud/docker/blob/940d58806c3e3d4409a7eee4859335e98139d09f/image.go#L218-225
func GenerateId() string {
	id := make([]byte, 32)
	_, err := io.ReadFull(rand.Reader, id)
	if err != nil {
		panic(err) // This shouldn't happen
	}
	return hex.EncodeToString(id)
}

var expandPathRegExp = regexp.MustCompile(`^~(\w*)`)

func ExpandPath(path string) string {
	matches := expandPathRegExp.FindAllStringSubmatch(path, -1)
	if len(matches) <= 0 {
		return path
	}

	replace := matches[0][0]
	username := matches[0][1]

	var usr *user.User
	var err error
	if username != "" {
		usr, err = user.Lookup(username)
	} else {
		usr, err = user.Current()
	}
	if err != nil {
		return path
	}

	homeDir := usr.HomeDir
	path = strings.Replace(path, replace, homeDir, 1)

	return path
}
