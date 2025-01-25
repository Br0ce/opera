package tool

import "net/url"

type Tool struct {
	Name        string
	Description string
	Parameters  Parameters
	Addr        url.URL
}

type Parameters struct {
	Properties map[string]any
	Required   []string
}

type ToolOption func(t *Tool)

func WithName(name string) ToolOption {
	return func(t *Tool) {
		t.Name = name
	}
}

func WithDescription(description string) ToolOption {
	return func(t *Tool) {
		t.Description = description
	}
}

func WithAddr(addr url.URL) ToolOption {
	return func(t *Tool) {
		t.Addr = addr
	}
}

func WithParameters(properties map[string]any, required []string) ToolOption {
	return func(t *Tool) {
		t.Parameters = Parameters{
			Properties: properties,
			Required:   required,
		}
	}
}

func MakeTool(options ...ToolOption) (Tool, error) {
	tool := &Tool{}
	for _, opt := range options {
		opt(tool)
	}
	return *tool, nil
}
