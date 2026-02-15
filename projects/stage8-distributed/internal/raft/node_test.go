package raft

import (
	"context"
	"testing"
)

func TestVoteAndAppend(t *testing.T) {
	n := NewNode(Config{ID: "n1", Peers: []string{"n1", "n2"}})
	reply := n.HandleRequestVote(context.Background(), RequestVoteArgs{Term: 1, CandidateID: "n2"})
	if !reply.VoteGranted {
		t.Fatal("expected vote granted")
	}
	app := n.HandleAppendEntries(context.Background(), AppendEntriesArgs{Term: 1, LeaderID: "n2", Entries: []LogEntry{{Index: 1, Term: 1}}})
	if !app.Success {
		t.Fatal("expected append success")
	}
}
