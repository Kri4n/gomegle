package cluster

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
	"strings"

	r "github.com/redis/go-redis/v9"
)

const (
	waitingListKey = "chat:waiting"
	nodeChanPrefix = "chat:node"
)

type RedisBroker struct {
	cli     *r.Client
	nodeID  string
	deliver func(Envelope)
	sub     *r.PubSub
}

func NewRedisBroker() (*RedisBroker, error) {
	url := os.Getenv("REDIS_URL")
	if url == "" {
		return nil, errors.New("REDIS_URL not set")
	}

	opt, err := r.ParseURL(url)
	if err != nil {
		return nil, err
	}

	nodeID := strings.TrimSpace(os.Getenv("NODE_ID"))
	if nodeID == "" {
		nodeID = randHex(8)
	}

	return &RedisBroker{
		cli: r.NewClient(opt),
		nodeID: nodeID,
	}, nil
}

func (b *RedisBroker) NodeID() string { return b.nodeID }
func (b *RedisBroker) OnDeliver(h func(Envelope)) { b.deliver = h }
func (b *RedisBroker) Start(ctx context.Context) error {
	// Subscribe to this node’s inbox channel.
	b.sub = b.cli.Subscribe(ctx, nodeChanPrefix+b.nodeID)

	go func() {
		ch := b.sub.Channel()
		for msg := range ch {
			env, err := decodeEnvelope([]byte(msg.Payload))
			if err == nil && b.deliver != nil {
				b.deliver(env)
			}
		}
	}()
	return  nil
}

func (b *RedisBroker) Stop(ctx context.Context) error {
	if b.sub != nil {
		return b.sub.Close()
	}
	return nil
}

// Try to pop someone from waiting; if none, push self and return nil.
func (b *RedisBroker) TryMatchOrEnqueue(ctx context.Context, clientID string) (*MatchResult, error) {
	// LPOP old partner; if empty, enqueue self.
	partnerID, err := b.cli.LPop(ctx, waitingListKey).Result()
	if err == r.Nil {
		// push self with TTL companion key to allow future cleanup if needed
		if err := b.cli.RPush(ctx, waitingListKey, encodeWaiting(clientID, b.nodeID)).Err(); err != nil {
			return nil, err
		}
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	id, node := decodeWaiting(partnerID)
	if id == "" || node == "" || id == clientID {
		// Fallback: enqueue self if bad data
		_ = b.cli.RPush(ctx, waitingListKey, encodeWaiting(clientID, b.nodeID)).Err()
		return nil, nil
	}
	return &MatchResult{PartnerClientID: id, PartnerNodeID: node}, nil
}

func (b *RedisBroker) RemoveFromQueue(ctx context.Context, clientID string) error {
	// Remove first match of this client from the waiting list.
	return b.cli.LRem(ctx, waitingListKey, 1, encodeWaiting(clientID, b.nodeID)).Err()
}

func (b *RedisBroker) SendToNode(ctx context.Context, nodeID string, env Envelope) error {
	data := encodeEnvelope(env)
	return b.cli.Publish(ctx, nodeChanPrefix+nodeID, data).Err()
}

// --- helpers 
func randHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func encodeWaiting(clientID, nodeID string) string {
	return clientID + "@" +	nodeID
}

func decodeWaiting(s string) (string, string) {
	parts := strings.SplitN(s, "@", 2)
	if len(parts) != 2 { return "", "" }
	return parts[0], parts[1]
}

// tiny, human-readable envelope format: type|to|from|hexpayload
func encodeEnvelope(e Envelope) string {
	return strings.Join([] string{e.Type, e.ToID, e.FromID, hex.EncodeToString(e.Payload)}, "|")
}

func decodeEnvelope(b []byte) (Envelope, error) {
	s := string(b)
	parts := strings.SplitN(s, "|", 4)
	if len(parts) != 4 { return Envelope{}, errors.New("bad envelope") }
	payload, err := hex.DecodeString(parts[3])
	if err != nil { return Envelope{}, err }
	return Envelope{ Type: parts[0], ToID: parts[1], FromID: parts[2], Payload: payload }, nil
}


