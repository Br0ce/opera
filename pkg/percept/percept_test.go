package percept

import (
	"reflect"
	"testing"

	"github.com/Br0ce/opera/pkg/tool"
	"github.com/Br0ce/opera/pkg/user"
)

func TestPercept_User(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		user  *user.Query
		want  user.Query
		want1 bool
	}{
		{
			name:  "ok",
			user:  &user.Query{},
			want:  user.Query{},
			want1: true,
		},
		{
			name:  "false",
			user:  nil,
			want:  user.Query{},
			want1: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p := Percept{
				user: test.user,
				tool: nil,
			}
			got, got1 := p.User()
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("Percept.User() got = %v, want %v", got, test.want)
			}
			if got1 != test.want1 {
				t.Errorf("Percept.User() got1 = %v, want %v", got1, test.want1)
			}
		})
	}
}

func TestPercept_Tool(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		tool  *tool.Response
		want  tool.Response
		want1 bool
	}{
		{
			name:  "ok",
			tool:  &tool.Response{},
			want:  tool.Response{},
			want1: true,
		},
		{
			name:  "false",
			tool:  nil,
			want:  tool.Response{},
			want1: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p := Percept{
				user: nil,
				tool: test.tool,
			}
			got, got1 := p.Tool()
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("Percept.Tool() got = %v, want %v", got, test.want)
			}
			if got1 != test.want1 {
				t.Errorf("Percept.Tool() got1 = %v, want %v", got1, test.want1)
			}
		})
	}
}
