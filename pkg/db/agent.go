package db

import "github.com/Br0ce/opera/pkg/agent"

type Agent interface {
	Add(agent agent.Agent) (string, error)
	Get(id string) (agent.Agent, error)
	Delete(id string) error
}
