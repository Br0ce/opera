package ids

import (
	"strings"

	"github.com/rs/xid"
)

const (
	agentPrefix = "age"
	seperator   = "-"
)

func UniqueAgent() string {
	return agentPrefix + seperator + unique()
}

func Valid(id string) bool {
	ii := strings.Split(id, "-")

	if len(ii) != 2 {
		return false
	}

	switch ii[0] {
	case agentPrefix:
		return valid(ii[1])
	default:
		return false
	}

}

func unique() string {
	return xid.New().String()
}

func valid(id string) bool {
	if _, err := xid.FromString(id); err != nil {
		return false
	}
	return true
}
