package raft

import "time"

// NodeState Raft 节点状态。
type NodeState int

const (
	Follower NodeState = iota
	Candidate
	Leader
)

// LogEntry 日志条目。
type LogEntry struct {
	Index   int64
	Term    int64
	Command []byte
}

// RequestVoteArgs 请求投票参数。
type RequestVoteArgs struct {
	Term         int64
	CandidateID  string
	LastLogIndex int64
	LastLogTerm  int64
}

// RequestVoteReply 请求投票响应。
type RequestVoteReply struct {
	Term        int64
	VoteGranted bool
}

// AppendEntriesArgs 追加日志参数。
type AppendEntriesArgs struct {
	Term         int64
	LeaderID     string
	PrevLogIndex int64
	PrevLogTerm  int64
	Entries      []LogEntry
	LeaderCommit int64
}

// AppendEntriesReply 追加日志响应。
type AppendEntriesReply struct {
	Term    int64
	Success bool
}

// Config 节点配置。
type Config struct {
	ID              string
	Peers           []string
	ElectionTimeout time.Duration
	Heartbeat       time.Duration
}
