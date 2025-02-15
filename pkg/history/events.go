package history

import (
	"time"

	"github.com/Br0ce/opera/pkg/tool"
	"github.com/Br0ce/opera/pkg/user"
)

type User struct {
	Content user.Query
	Created time.Time
}

type Assistant struct {
	Content string
	Created time.Time
}

type ToolCalls struct {
	Content []tool.Call
	Created time.Time
}

type ToolResponse struct {
	Content tool.Response
	Created time.Time
}

type System struct {
	Content string
	Created time.Time
}
