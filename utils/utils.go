package utils

import (
	"crypto/rand"
	"encoding/hex"
	"io"
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
