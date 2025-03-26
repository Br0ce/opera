package tool

import (
	"fmt"
	"net/url"
)

// Tool represents a remote tool service and all data needed to call it.
type Tool struct {
	name        string
	description string
	parameters  Parameters
	addr        url.URL
}

type Parameters struct {
	Properties map[string]any
	Required   []string
}

type Option func(t *Tool)

func WithName(name string) Option {
	return func(t *Tool) {
		t.name = name
	}
}

func WithDescription(description string) Option {
	return func(t *Tool) {
		t.description = description
	}
}

func WithAddr(addr url.URL) Option {
	return func(t *Tool) {
		t.addr = addr
	}
}

func WithParameters(properties map[string]any, required []string) Option {
	return func(t *Tool) {
		t.parameters = Parameters{
			Properties: properties,
			Required:   required,
		}
	}
}

func MakeTool(options ...Option) (Tool, error) {
	tool := &Tool{}
	for _, opt := range options {
		opt(tool)
	}
	if tool.name == "" {
		return Tool{}, fmt.Errorf("name invalid")
	}
	if tool.description == "" {
		return Tool{}, fmt.Errorf("description invalid")
	}
	if tool.addr.Host == "" {
		return Tool{}, fmt.Errorf("addr invalid")
	}
	if tool.parameters.Properties == nil {
		return Tool{}, fmt.Errorf("Parameters.Properties invalid")
	}
	return *tool, nil
}

func (t Tool) Name() string {
	return t.name
}

func (t Tool) Description() string {
	return t.description
}

func (t Tool) Addr() url.URL {
	return t.addr
}

func (t Tool) Parameters() Parameters {
	return t.parameters
}
