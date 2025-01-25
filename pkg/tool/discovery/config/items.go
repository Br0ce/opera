package config

import (
	"fmt"
	"net/url"

	"github.com/Br0ce/opera/pkg/tool"
)

type Items struct {
	Tools []Item `json:"tools"`
}

type Item struct {
	Name        string     `json:"Name"`
	Description string     `json:"Description"`
	Parameters  Parameters `json:"Parameters"`
	Addr        string     `json:"Addr"`
}

type Parameters struct {
	Properties map[string]any `json:"Properties"`
	Required   []string       `json:"Required"`
}

func (i Item) Decode() (tool.Tool, error) {
	addr, err := url.Parse(i.Addr)
	if err != nil {
		return tool.Tool{}, fmt.Errorf("parse addr: %w", err)
	}
	return tool.MakeTool(
		tool.WithName(i.Name),
		tool.WithDescription(i.Description),
		tool.WithParameters(i.Parameters.Properties, i.Parameters.Required),
		tool.WithAddr(*addr))
}

func (p Parameters) Decode() tool.Parameters {
	// todo Need to validate parameters.
	return tool.Parameters{
		Properties: p.Properties,
		Required:   p.Required,
	}
}
