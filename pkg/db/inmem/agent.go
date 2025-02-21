package inmem

import (
	"sync"

	"github.com/Br0ce/opera/pkg/agent"
	"github.com/Br0ce/opera/pkg/db"
	"github.com/Br0ce/opera/pkg/ids"
)

var _ db.Agent = (*Agent)(nil)

type Agent struct {
	agents sync.Map
}

func NewAgentDB() *Agent {
	return &Agent{}
}

// Add stores the Agent and returns the id for which the Agent can be retrieved.
func (ag *Agent) Add(agent agent.Agent) (string, error) {
	id := ids.UniqueAgent()
	_, ok := ag.agents.LoadOrStore(id, agent)
	if ok {
		// TODO check if needed
		return "", db.ErrAlreadyExists
	}
	return id, nil
}

// Get returns the Agent stored for the given id.
// If no Agent is found for the given id, a db.ErrNotFound is returned.
func (ag *Agent) Get(id string) (agent.Agent, error) {
	if id == "" {
		return agent.Agent{}, db.ErrInvalidID
	}
	v, ok := ag.agents.Load(id)
	if !ok {
		return agent.Agent{}, db.ErrNotFound
	}

	a, ok := v.(agent.Agent)
	if !ok {
		// This should not happen.
		return agent.Agent{}, db.ErrInternal
	}
	return a, nil
}

func (ag *Agent) Update(id string, agent agent.Agent) error {
	_, ok := ag.agents.LoadOrStore(id, agent)
	if !ok {
		return db.ErrNotFound
	}
	return nil
}

// Delete deletes the Agent stored for the given id.
func (ag *Agent) Delete(id string) error {
	ag.agents.Delete(id)
	return nil
}
