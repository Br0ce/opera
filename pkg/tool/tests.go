package tool

import "net/url"

func TestToolA() Tool {
	return Tool{
		name:        "get_names",
		description: "Get all names from my db for the given location.",
		parameters: Parameters{
			Properties: map[string]any{
				"location": map[string]any{
					"type": "string",
				},
			},
			Required: []string{"location"},
		},
		addr: url.URL{Host: "mySvc"},
	}
}

func TestToolB() Tool {
	return Tool{
		name:        "get_numbers",
		description: "Get all numbers from my db for the given location.",
		parameters: Parameters{
			Properties: map[string]any{
				"location": map[string]any{
					"type": "string",
				},
			},
			Required: []string{"location"},
		},
		addr: url.URL{Host: "myOtherSvc"},
	}
}

func TestTools() []Tool {
	return []Tool{TestToolA(), TestToolB()}
}
