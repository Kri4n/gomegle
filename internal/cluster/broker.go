package cluster

import "context"

// MatchResult is returned after trying to pair via the global queue.
type MatchResult struct {
	PartnerClientID string
	PartnerNodeID   string
}

// Envelope is what we publish between nodes.
type Envelope struct {
	Type string // "msg" | "notify" | "unpair"
	ToID string // recipient client id (local to this node)
	FromID string
	Payload []byte
}

// Broker hides the Redis integration so the Hub can be agnostic.
type Broker interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	NodeID() string

	// Global waiting room
	TryMatchOrEnqueue(ctx context.Context, clientID string) (*MatchResult, error)
	RemoveFromQueue(ctx context.Context, clientID string) error

	// Cross-node delivery to a node’s inbox
	SendToNode(ctx context.Context, nodeID string, env Envelope) error

	// Register a local delivery handler for this node
	OnDeliver(func(env Envelope))
}