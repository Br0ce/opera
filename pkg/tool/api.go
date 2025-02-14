package tool

// Call provides the content for a tool action.
type Call struct {
	ID string
	// The name of the tool to call.
	Name      string
	Arguments string
}

// Response containes the output of the called tool.
type Response struct {
	ID      string
	Content string
}
