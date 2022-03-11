package plans

import (
	"sync"

	"github.com/hashicorp/terraform/internal/addrs"
)

type Conditions map[string]*ConditionResult

type ConditionResult struct {
	Address      addrs.Checkable
	Result       bool
	Unknown      bool
	Type         ConditionType
	ErrorMessage string
}

func NewConditions() Conditions {
	return make(Conditions)
}

func (c Conditions) SyncWrapper() *ConditionsSync {
	return &ConditionsSync{
		results: c,
	}
}

type ConditionsSync struct {
	lock    sync.Mutex
	results Conditions
}

func (cs *ConditionsSync) SetResult(addr string, result *ConditionResult) {
	if cs == nil {
		panic("SetResult on nil Conditions")
	}
	cs.lock.Lock()
	defer cs.lock.Unlock()

	cs.results[addr] = result
}
