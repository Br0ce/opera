package action

import (
	"reflect"
	"testing"

	"github.com/Br0ce/opera/pkg/tool"
)

func TestMakeUser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
		want    Action
	}{
		{
			name:    "pass",
			content: "The weather is fine today",
			want: Action{
				user: "The weather is fine today",
			},
		},
		{
			name:    "empty content",
			content: "",
			want: Action{
				user: "",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := MakeUser(test.content); !reflect.DeepEqual(got, test.want) {
				t.Errorf("MakeUser() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestMakeTool(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		calls  []tool.Call
		reason string
		want   Action
	}{
		{
			name: "pass",
			calls: []tool.Call{{
				ID:        "123",
				Name:      "get_weather",
				Arguments: "this is a JSON string",
			}},
			reason: "my reason",
			want: Action{
				tool: []tool.Call{{
					ID:        "123",
					Name:      "get_weather",
					Arguments: "this is a JSON string",
				}},
				reason: "my reason",
			},
		},
		{
			name: "pass without reason",
			calls: []tool.Call{{
				ID:        "123",
				Name:      "get_weather",
				Arguments: "this is a JSON string",
			}},
			reason: "",
			want: Action{
				tool: []tool.Call{{
					ID:        "123",
					Name:      "get_weather",
					Arguments: "this is a JSON string",
				}},
				reason: "",
			},
		},
		{
			name: "two calls",
			calls: []tool.Call{{
				ID:        "123",
				Name:      "get_weather",
				Arguments: "this is a JSON string",
			},
				{
					ID:        "456",
					Name:      "get_shark_level",
					Arguments: "this is a JSON string",
				},
			},
			want: Action{
				tool: []tool.Call{{
					ID:        "123",
					Name:      "get_weather",
					Arguments: "this is a JSON string",
				},
					{
						ID:        "456",
						Name:      "get_shark_level",
						Arguments: "this is a JSON string",
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := MakeTool(test.calls, test.reason); !reflect.DeepEqual(got, test.want) {
				t.Errorf("MakeTool() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestAction_User(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
		tool    []tool.Call
		want    string
		ok      bool
	}{
		{
			name:    "user",
			content: "this is some content",
			want:    "this is some content",
			tool:    nil,
			ok:      true,
		},
		{
			name: "tool",
			tool: []tool.Call{},
			ok:   false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			a := Action{
				user: test.content,
				tool: test.tool,
			}
			got, ok := a.User()
			if got != test.want {
				t.Errorf("Action.User() got = %v, want %v", got, test.want)
			}
			if ok != test.ok {
				t.Errorf("Action.User() ok = %v, want %v", ok, test.ok)
			}
		})
	}
}

func TestAction_Tool(t *testing.T) {
	t.Parallel()

	type fields struct {
		tool []tool.Call
	}
	tests := []struct {
		name   string
		fields fields
		want   []tool.Call
		ok     bool
	}{
		{
			name: "tool",
			fields: fields{
				tool: []tool.Call{{
					ID:        "1234",
					Name:      "MyFunc",
					Arguments: "json",
				}},
			},
			want: []tool.Call{{
				ID:        "1234",
				Name:      "MyFunc",
				Arguments: "json",
			}},
			ok: true,
		},
		{
			name:   "user",
			fields: fields{},
			want:   nil,
			ok:     false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			a := Action{
				tool: test.fields.tool,
			}
			got, ok := a.Tool()
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("Action.Tool() got = %v, want %v", got, test.want)
			}
			if ok != test.ok {
				t.Errorf("Action.Tool() ok = %v, want %v", ok, test.ok)
			}
		})
	}
}

func TestAction_Reason(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		reason     string
		wantReason string
		wantOk     bool
	}{
		{
			name:       "pass",
			reason:     "my reason",
			wantReason: "my reason",
			wantOk:     true,
		},
		{
			name:       "empty reason",
			reason:     "",
			wantReason: "",
			wantOk:     false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			a := Action{
				reason: test.reason,
			}
			gotReason, gotOk := a.Reason()
			if gotReason != test.wantReason {
				t.Errorf("Action.Reason() gotReason = %v, want %v", gotReason, test.wantReason)
			}
			if gotOk != test.wantOk {
				t.Errorf("Action.Reason() gotOk = %v, want %v", gotOk, test.wantOk)
			}
		})
	}
}
