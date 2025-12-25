package control

import (
	"encoding/json"
	"io"
	"razpravljalnica/internal/api"
	"sync"

	"github.com/hashicorp/raft"
)

type StateLog struct {
	HealthyNodes []*api.NodeInfo
	PendingNode  *api.NodeInfo
	DeadNodes    []*api.NodeInfo
}

type State struct {
	lock sync.RWMutex
	StateLog
}

var _ raft.FSM = (*State)(nil)

func NewState(initial StateLog) *State {
	return &State{
		StateLog: initial,
	}
}

func (s *State) Apply(log *raft.Log) any {
	var stateLog StateLog
	if err := json.Unmarshal(log.Data, &stateLog); err != nil {
		return err
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	s.HealthyNodes = stateLog.HealthyNodes
	s.PendingNode = stateLog.PendingNode
	s.DeadNodes = stateLog.DeadNodes

	return nil
}

func (s *State) Restore(snapshot io.ReadCloser) error {
	defer snapshot.Close()

	var newState State
	if err := json.NewDecoder(snapshot).Decode(&newState); err != nil {
		return err
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	s.HealthyNodes = newState.HealthyNodes
	s.PendingNode = newState.PendingNode
	s.DeadNodes = newState.DeadNodes

	return nil
}

func (s *State) Snapshot() (raft.FSMSnapshot, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	healthyCopy := make([]*api.NodeInfo, len(s.HealthyNodes))
	copy(healthyCopy, s.HealthyNodes)

	deadCopy := make([]*api.NodeInfo, len(s.DeadNodes))
	copy(deadCopy, s.DeadNodes)

	snap := &fsmSnapshot{
		HealthyNodes: healthyCopy,
		PendingNode:  s.PendingNode,
		DeadNodes:    deadCopy,
	}

	return snap, nil
}

type fsmSnapshot struct {
	HealthyNodes []*api.NodeInfo `json:"healthy_nodes"`
	PendingNode  *api.NodeInfo   `json:"pending_node"`
	DeadNodes    []*api.NodeInfo `json:"dead_nodes"`
}

var _ raft.FSMSnapshot = (*fsmSnapshot)(nil)

func (f *fsmSnapshot) Persist(sink raft.SnapshotSink) error {
	enc := json.NewEncoder(sink)

	if err := enc.Encode(f); err != nil {
		sink.Cancel()
		return err
	}

	return sink.Close()
}

func (f *fsmSnapshot) Release() {}
