package percept

const (
	user      = "user"
	tool      = "tool"
	TextType  = "text"
	ImageType = "image"
)

type Percept struct {
	pType string
	user  User
	tool  Tool
}

type User struct {
	Type     string
	Text     string
	ImageUrl string
}

type Tool struct {
	ID      string
	Content string
}

func MakeTextUser(content string) Percept {
	return Percept{
		pType: user,
		user: User{
			Type: TextType,
			Text: content,
		},
	}
}

func MakeTool(callID string, content string) Percept {
	return Percept{
		pType: tool,
		tool: Tool{
			ID:      callID,
			Content: content,
		},
	}
}

func (p Percept) User() ([]User, bool) {
	if p.pType != user {
		return nil, false
	}
	return []User{p.user}, true
}

func (p Percept) Tool() (Tool, bool) {
	if p.pType != tool {
		return Tool{}, false
	}
	return p.tool, true
}
