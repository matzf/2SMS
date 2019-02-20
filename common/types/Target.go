package types

import (
	"fmt"
)

type Target struct {
	Name   string            `json:"name,omitempty"`
	ISD    string            `json:"isd,omitempty"`
	AS     string            `json:"as,omitempty"`
	IP     string            `json:"ip,omitempty"`
	Port   string            `json:"port,omitempty"`
	Path   string            `json:"path,omitempty"`
	Labels map[string]string `json:"labels,omitempty"`
}

func (t *Target) BuildJobName() string {
	return fmt.Sprint(t.ISD) + "-" + fmt.Sprint(t.AS) + " " + t.IP + " " + t.Name
}
