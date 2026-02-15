package raft

import (
	"context"
	"sync"
	"time"
)

// Node 简化 Raft 节点。
type Node struct {
	id          string
	peers       []string
	state       NodeState
	currentTerm int64
	votedFor    string
	leader      string
	log         []LogEntry
	commitIndex int64
	lastApplied int64
	store       map[string]string
	mu          sync.RWMutex
}

// NewNode 创建节点。
func NewNode(cfg Config) *Node {
	if cfg.ElectionTimeout <= 0 {
		cfg.ElectionTimeout = 300 * time.Millisecond
	}
	if cfg.Heartbeat <= 0 {
		cfg.Heartbeat = 100 * time.Millisecond
	}
	return &Node{id: cfg.ID, peers: cfg.Peers, state: Follower, log: make([]LogEntry, 0), store: make(map[string]string)}
}

// BecomeLeader 切换 Leader。
func (n *Node) BecomeLeader() {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.state = Leader
	n.leader = n.id
}

// HandleRequestVote 处理投票。
func (n *Node) HandleRequestVote(ctx context.Context, args RequestVoteArgs) RequestVoteReply {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return RequestVoteReply{Term: n.currentTerm, VoteGranted: false}
	}

	n.mu.Lock()
	defer n.mu.Unlock()
	if args.Term < n.currentTerm {
		return RequestVoteReply{Term: n.currentTerm, VoteGranted: false}
	}
	if args.Term > n.currentTerm {
		n.currentTerm = args.Term
		n.votedFor = ""
		n.state = Follower
	}
	if n.votedFor == "" || n.votedFor == args.CandidateID {
		n.votedFor = args.CandidateID
		return RequestVoteReply{Term: n.currentTerm, VoteGranted: true}
	}
	return RequestVoteReply{Term: n.currentTerm, VoteGranted: false}
}

// HandleAppendEntries 处理日志追加。
func (n *Node) HandleAppendEntries(ctx context.Context, args AppendEntriesArgs) AppendEntriesReply {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return AppendEntriesReply{Term: n.currentTerm, Success: false}
	}

	n.mu.Lock()
	defer n.mu.Unlock()
	if args.Term < n.currentTerm {
		return AppendEntriesReply{Term: n.currentTerm, Success: false}
	}
	n.currentTerm = args.Term
	n.state = Follower
	n.leader = args.LeaderID
	if len(args.Entries) > 0 {
		n.log = append(n.log, args.Entries...)
		n.commitIndex = int64(len(n.log))
	}
	return AppendEntriesReply{Term: n.currentTerm, Success: true}
}

// ApplyKV 写入状态机。
func (n *Node) ApplyKV(ctx context.Context, key, value string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	n.store[key] = value
	n.lastApplied++
	return nil
}

// GetKV 读取状态机。
func (n *Node) GetKV(key string) (string, bool) {
	n.mu.RLock()
	defer n.mu.RUnlock()
	v, ok := n.store[key]
	return v, ok
}

// State 返回节点状态。
func (n *Node) State() NodeState {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.state
}
