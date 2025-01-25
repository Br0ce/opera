package message

import (
	"time"

	"github.com/Br0ce/opera/pkg/action"
	"github.com/Br0ce/opera/pkg/percept"
)

const (
	UserRole      = "user"
	AssistantRole = "assistant"
	ToolRole      = "tool"
	SystemRole    = "system"
)

type Message struct {
	Role      string
	Calls     []action.Call
	User      []percept.User
	Tool      percept.Tool
	Assistent string
	System    string
	Created   time.Time
}

func (m Message) ForUser() bool {
	return len(m.Calls) == 0
}
