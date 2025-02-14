package percept

import (
	"github.com/Br0ce/opera/pkg/tool"
	"github.com/Br0ce/opera/pkg/user"
)

// Percept is a standardized container for all possible inputs to an agent.
// Perceptions can be of type user or of type tool.
// A user perception represent usually the initial question to the agent. Tool
// perceptions hold the response of a tool call.
type Percept struct {
	user *user.Query
	tool *tool.Response
}

func MakeUser(query user.Query) Percept {
	return Percept{
		user: &query,
	}
}

func MakeTool(callID string, content string) Percept {
	return Percept{
		tool: &tool.Response{
			ID:      callID,
			Content: content,
		},
	}
}

// User reports if the percept is of type user. If true the user.Query is returned.
func (p Percept) User() (user.Query, bool) {
	if p.user == nil {
		return user.Query{}, false
	}
	return *p.user, true
}

// Tool reports if the percept is of type tool. If true the tool.Response is returned.
func (p Percept) Tool() (tool.Response, bool) {
	if p.tool == nil {
		return tool.Response{}, false
	}
	return *p.tool, true
}
