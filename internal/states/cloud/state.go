package cloud

import (
	"github.com/hashicorp/terraform/internal/states"
	"github.com/hashicorp/terraform/internal/states/remote"
	"github.com/hashicorp/terraform/internal/states/statemgr"
)

// State is similar to remote State and delegates to it, except in the case of output values,
// which use a separate methodology that ensures the caller is authorized to read cloud
// workspace outputs.
type State struct {
	Client remote.Client

	delegate remote.State
}

// Proof that cloud State is a statemgr.Persistent interface
var _ statemgr.Persistent = (*State)(nil)

func NewState(client remote.Client) *State {
	return &State{
		Client:   client,
		delegate: remote.State{Client: client},
	}
}

// State delegates calls to read State to the remote State
func (s *State) State() *states.State {
	return s.delegate.State()
}

// RefreshState delegates calls to refresh State to the remote State
func (s *State) RefreshState() error {
	return s.delegate.RefreshState()
}

// RefreshState delegates calls to refresh State to the remote State
func (s *State) PersistState() error {
	return s.delegate.PersistState()
}

// GetOutputValues
func (s *State) GetOutputValues() (map[string]*states.OutputValue, error) {

}
