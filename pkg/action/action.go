package action

const (
	forUser = "user"
	forTool = "call"
)

// Action is a type to hold the content needed for an upcoming action.
// Actions can be of type 'tool' or 'user', indicating if the action is
// meant to be executed by a tool or a user.
type Action struct {
	aType string
	user  string
	tool  []Call
}

// Call provides the content for a tool action.
type Call struct {
	ID string
	// The name of the tool.
	Name      string
	Arguments string
}

func MakeUser(content string) Action {
	return Action{
		aType: forUser,
		user:  content,
	}
}

func MakeTool(calls []Call) Action {
	return Action{
		aType: forTool,
		tool:  calls,
	}
}

// User reports if the action is of type user. If true the content for the user action is returned.
func (a Action) User() (content string, ok bool) {
	if a.aType != forUser {
		return "", false
	}
	return a.user, true
}

// Tool reports if the action is of type tool. If true a slice of calls to the tool services is returned.
func (a Action) Tool() (content []Call, ok bool) {
	if a.aType != forTool {
		return nil, false
	}
	return a.tool, true
}
