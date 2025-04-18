package action

import "github.com/Br0ce/opera/pkg/tool"

// Action is a type to hold the content needed for an upcoming action.
// An Action can be of type 'tool' or 'user', e.q. it is meant to be
// executed by a tool or a user.
type Action struct {
	user string
	// The reason for the action.
	reason string
	tool   []tool.Call
}

func MakeUser(content string) Action {
	return Action{
		user: content,
	}
}

func MakeTool(calls []tool.Call, reason string) Action {
	return Action{
		reason: reason,
		tool:   calls,
	}
}

// User reports if the action is of type user. If true the content for the user action is returned.
func (a Action) User() (content string, ok bool) {
	if a.user == "" {
		return "", false
	}
	return a.user, true
}

// Tool reports if the action is of type tool. If true a slice of calls to the tool services is returned.
func (a Action) Tool() (content []tool.Call, ok bool) {
	if a.tool == nil {
		return nil, false
	}
	return a.tool, true
}

// Reason reports if the action provides a reason. If true the reason is returned.
func (a Action) Reason() (reason string, ok bool) {
	if a.reason == "" {
		return "", false
	}
	return a.reason, true
}
