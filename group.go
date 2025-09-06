package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/Jille/raft-grpc-leader-rpc/leaderhealth"
	transport "github.com/Jille/raft-grpc-transport"
	"github.com/hashicorp/raft"
	boltdb "github.com/hashicorp/raft-boltdb"
	"google.golang.org/grpc"
)

type GroupManager struct {
	mu     sync.RWMutex
	groups map[string]*GroupInstance
}

type GroupInstance struct {
	GroupID     string
	Raft        *raft.Raft
	Transport   *transport.Manager
	WordTracker *wordTracker
}

func NewGroupManager() *GroupManager {
	return &GroupManager{
		groups: make(map[string]*GroupInstance),
	}
}

func (gm *GroupManager) CreateGroup(ctx context.Context, groupID, nodeID, address string, bootstrap bool) (*GroupInstance, error) {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	if _, exists := gm.groups[groupID]; exists {
		return nil, fmt.Errorf("group %s already exists", groupID)
	}

	wt := &wordTracker{}

	r, tm, err := NewRaftWithGroupID(ctx, groupID, nodeID, address, wt, bootstrap)
	if err != nil {
		return nil, fmt.Errorf("failed to create Raft for group %s: %v", groupID, err)
	}

	instance := &GroupInstance{
		GroupID:     groupID,
		Raft:        r,
		Transport:   tm,
		WordTracker: wt,
	}

	gm.groups[groupID] = instance
	log.Printf("Created Raft group: %s (node: %s)", groupID, nodeID)
	return instance, nil
}

func (gm *GroupManager) GetGroup(groupID string) (*GroupInstance, error) {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	group, exists := gm.groups[groupID]
	if !exists {
		return nil, fmt.Errorf("group %s not found", groupID)
	}
	return group, nil
}

func (gm *GroupManager) RegisterWithGRPC(s *grpc.Server) {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	for _, group := range gm.groups {
		group.Transport.Register(s)
		leaderhealth.Setup(group.Raft, s, []string{fmt.Sprintf("Example-%s", group.GroupID)})
	}
}

func NewRaftWithGroupID(ctx context.Context, groupID, nodeID, address string, fsm raft.FSM, bootstrap bool) (*raft.Raft, *transport.Manager, error) {
	c := raft.DefaultConfig()
	c.LocalID = raft.ServerID(nodeID)

	baseDir := filepath.Join(*raftDir, groupID, nodeID)

	ldb, err := boltdb.NewBoltStore(filepath.Join(baseDir, "logs.dat"))
	if err != nil {
		return nil, nil, fmt.Errorf(`boltdb.NewBoltStore(%q): %v`, filepath.Join(baseDir, "logs.dat"), err)
	}

	sdb, err := boltdb.NewBoltStore(filepath.Join(baseDir, "stable.dat"))
	if err != nil {
		return nil, nil, fmt.Errorf(`boltdb.NewBoltStore(%q): %v`, filepath.Join(baseDir, "stable.dat"), err)
	}

	fss, err := raft.NewFileSnapshotStore(baseDir, 3, os.Stderr)
	if err != nil {
		return nil, nil, fmt.Errorf(`raft.NewFileSnapshotStore(%q, ...): %v`, baseDir, err)
	}

	tm := transport.New(raft.ServerAddress(address), []grpc.DialOption{grpc.WithInsecure()})

	r, err := raft.NewRaft(c, fsm, ldb, sdb, fss, tm.Transport())
	if err != nil {
		return nil, nil, fmt.Errorf("raft.NewRaft: %v", err)
	}

	if bootstrap {
		cfg := raft.Configuration{
			Servers: []raft.Server{
				{
					Suffrage: raft.Voter,
					ID:       raft.ServerID(nodeID),
					Address:  raft.ServerAddress(address),
				},
			},
		}
		f := r.BootstrapCluster(cfg)
		if err := f.Error(); err != nil {
			return nil, nil, fmt.Errorf("raft.Raft.BootstrapCluster: %v", err)
		}
	}

	return r, tm, nil
}

