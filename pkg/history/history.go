package history

import (
	"iter"
	"time"

	"github.com/Br0ce/opera/pkg/action"
	"github.com/Br0ce/opera/pkg/percept"
)

type History struct {
	events []any
}

func (h *History) AddSystem(prompt string) {
	// An empty string is a valid system prompt.
	system := System{
		Content: prompt,
		Created: time.Now().UTC(),
	}
	h.events = append(h.events, system)
}

func (h *History) AddPercepts(percepts []percept.Percept) {
	h.events = append(h.events, events(percepts)...)
}

func (h *History) AddAction(action action.Action) {
	if content, ok := action.User(); ok {
		assist := Assistant{
			Content: content,
			Created: time.Now().UTC(),
		}
		h.events = append(h.events, assist)
		return
	}
	if content, ok := action.Tool(); ok {
		calls := ToolCalls{
			Content: content,
			Created: time.Now().UTC(),
		}
		h.events = append(h.events, calls)
	}
}

func (h *History) All() iter.Seq2[int, any] {
	return func(yield func(int, any) bool) {
		for i, e := range h.events {
			if !yield(i, e) {
				return
			}
		}
	}
}

// events returns the given perceptions as a slice of Events.
func events(percepts []percept.Percept) []any {
	ee := make([]any, 0, len(percepts))
	for _, percept := range percepts {
		if u, ok := percept.User(); ok {
			u := User{
				Content: u,
				Created: time.Now().UTC(),
			}
			ee = append(ee, u)
			continue
		}

		if t, ok := percept.Tool(); ok {
			to := ToolResponse{
				Content: t,
				Created: time.Now().UTC(),
			}
			ee = append(ee, to)
		}
	}
	return ee
}
