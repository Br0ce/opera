package action

import (
	"reflect"
	"testing"
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
				aType: forUser,
				user:  "The weather is fine today",
			},
		},
		{
			name:    "empty content",
			content: "",
			want: Action{
				aType: forUser,
				user:  "",
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
		name  string
		calls []Call
		want  Action
	}{
		{
			name: "pass",
			calls: []Call{{
				ID:        "123",
				Name:      "get_weather",
				Arguments: "this is a JSON string",
			}},
			want: Action{
				aType: forTool,
				tool: []Call{{
					ID:        "123",
					Name:      "get_weather",
					Arguments: "this is a JSON string",
				}},
			},
		},
		{
			name: "two calls",
			calls: []Call{{
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
				aType: forTool,
				tool: []Call{{
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
			if got := MakeTool(test.calls); !reflect.DeepEqual(got, test.want) {
				t.Errorf("MakeTool() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestAction_User(t *testing.T) {
	t.Parallel()

	type fields struct {
		aType string
		user  string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
		ok     bool
	}{
		{
			name: "user",
			fields: fields{
				aType: forUser,
				user:  "this is some content",
			},
			want: "this is some content",
			ok:   true,
		},
		{
			name: "tool",
			fields: fields{
				aType: forTool,
			},
			want: "",
			ok:   false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			a := Action{
				aType: test.fields.aType,
				user:  test.fields.user,
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
		aType string
		tool  []Call
	}
	tests := []struct {
		name   string
		fields fields
		want   []Call
		ok     bool
	}{
		{
			name: "tool",
			fields: fields{
				aType: forTool,
				tool: []Call{{
					ID:        "1234",
					Name:      "MyFunc",
					Arguments: "json",
				}},
			},
			want: []Call{{
				ID:        "1234",
				Name:      "MyFunc",
				Arguments: "json",
			}},
			ok: true,
		},
		{
			name: "user",
			fields: fields{
				aType: forUser,
			},
			want: nil,
			ok:   false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			a := Action{
				aType: test.fields.aType,
				tool:  test.fields.tool,
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
