package main

import (
	"fmt"
	"strings"
	"github.com/gocardless/rig"
)

type DescriptorResolver struct {
	str string
	dir string
}

func NewDescriptorResolver(str string, dir string) *DescriptorResolver {
	return &DescriptorResolver{str: str, dir: dir}
}

func (r *DescriptorResolver) GetDescriptor() (*rig.Descriptor, error) {
	var d rig.Descriptor

	parts := strings.SplitN(r.str, ":", 3)
	if parts[0] == "" {
		return nil, fmt.Errorf("No stack specified")
	}
	d.Stack = parts[0]

	if len(parts) > 1 && parts[1] != "" {
		d.Service = parts[1]
	}

	if len(parts) > 2 && parts[2] != "" {
		d.Process = parts[2]
	}

	return &d, nil
}
