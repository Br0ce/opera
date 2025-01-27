package tool

import (
	"net/url"
	"reflect"
	"testing"
)

func TestMakeTool(t *testing.T) {
	tests := []struct {
		name    string
		want    Tool
		options []ToolOption
		wantErr bool
	}{
		{
			name: "minimal pass",
			want: Tool{
				name:        "MyName",
				description: "My description",
				addr:        url.URL{Host: "MyHost"},
				parameters: Parameters{
					Properties: map[string]any{
						"key": "value",
					},
					Required: []string{"value"},
				},
			},
			options: []ToolOption{
				WithName("MyName"),
				WithDescription("My description"),
				WithAddr(url.URL{Host: "MyHost"}),
				WithParameters(map[string]any{
					"key": "value",
				}, []string{"value"}),
			},
			wantErr: false,
		},
		{
			name: "empty name",
			options: []ToolOption{
				WithName(""),
				WithAddr(url.URL{Host: "MyHost"}),
				WithParameters(map[string]any{
					"key": "value",
				}, []string{"value"}),
			},
			wantErr: true,
		},
		{
			name: "empty description",
			options: []ToolOption{
				WithName("MyName"),
				WithDescription(""),
				WithAddr(url.URL{Host: "MyHost"}),
				WithParameters(map[string]any{
					"key": "value",
				}, []string{"value"}),
			},
			wantErr: true,
		},
		{
			name: "empty addr",
			options: []ToolOption{
				WithName("MyName"),
				WithDescription("My description"),
				WithAddr(url.URL{}),
				WithParameters(map[string]any{
					"key": "value",
				}, []string{"value"}),
			},
			wantErr: true,
		},
		{
			name: "nil propertires",
			options: []ToolOption{
				WithName("MyName"),
				WithDescription("My description"),
				WithAddr(url.URL{Host: "MyHost"}),
				WithParameters(nil, []string{"value"}),
			},
			wantErr: true,
		},
		{
			name: "no name",
			options: []ToolOption{
				WithDescription("My description"),
				WithAddr(url.URL{Host: "MyHost"}),
				WithParameters(map[string]any{
					"key": "value",
				}, []string{"value"}),
			},
			wantErr: true,
		},
		{
			name: "no description",
			options: []ToolOption{
				WithName("MyName"),
				WithAddr(url.URL{Host: "MyHost"}),
				WithParameters(map[string]any{
					"key": "value",
				}, []string{"value"}),
			},
			wantErr: true,
		},
		{
			name: "no addr",
			options: []ToolOption{
				WithName("MyName"),
				WithDescription("My description"),
				WithParameters(map[string]any{
					"key": "value",
				}, []string{"value"}),
			},
			wantErr: true,
		},
		{
			name: "no parameters",
			options: []ToolOption{
				WithName("MyName"),
				WithAddr(url.URL{Host: "MyHost"}),
				WithDescription("My description"),
			},
			wantErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := MakeTool(test.options...)
			if (err != nil) != test.wantErr {
				t.Errorf("MakeTool() error = %v, wantErr %v", err, test.wantErr)
				return
			}
			if test.wantErr {
				return
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("MakeTool() = %v, want %v", got, test.want)
			}
		})
	}
}
