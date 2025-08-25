package server

import (
	"context"
	"gomegle/internal/cluster"
	"sync"
)

type Hub struct {
	Register    chan *Client
	Unregister  chan *Client
	Waiting     []*Client
	clientsByID map[string]*Client
	Stats       *Stats

	// cluster mode (nil in standalone)
	broker cluster.Broker
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.RWMutex
}

type Stats struct {
	Connected int
	Waiting   int
	Paired    int
}

func NewHub() *Hub {
	ctx, cancel := context.WithCancel(context.Background())
	return &Hub{
		Register:    make(chan *Client),
		Unregister:  make(chan *Client),
		Waiting:     make([]*Client, 0, 64),
		clientsByID: make(map[string]*Client),
		Stats:       &Stats{},
		ctx:         ctx,
		cancel:      cancel,
	}
}

func (h *Hub) EnableCluster(b cluster.Broker) error {
	h.broker = b
	b.OnDeliver(func(env cluster.Envelope) {
		// Received a cross-node delivery; route to local client
		h.mu.RLock()
		c := h.clientsByID[env.ToID]
		h.mu.RUnlock()
		if c != nil {
			select {
			case c.Send <- env.Payload:
			default:
			}
		}
	})
	return b.Start(h.ctx)
}

func (h *Hub) Stop() { h.cancel() }

func (h *Hub) Run() {
	for {
		select {
		case c := <-h.Register:
			h.mu.Lock()
			h.clientsByID[c.ID] = c
			h.mu.Unlock()
			h.Stats.Connected++

			if h.broker == nil {
				// standalone local pairing
				if len(h.Waiting) > 0 {
					partner := h.Waiting[0]
					h.Waiting = h.Waiting[1:]
					pair(c, partner)
					h.Stats.Paired++
					if h.Stats.Waiting > 0 {
						h.Stats.Waiting--
					}
				} else {
					h.Waiting = append(h.Waiting, c)
					h.Stats.Waiting++
				}
				break
			}

			// cluster pairing via Redis
			if res, err := h.broker.TryMatchOrEnqueue(h.ctx, c.ID); err == nil && res != nil {
				// Matched immediately
				if res.PartnerNodeID == c.NodeID {
					// partner is local
					h.mu.RLock()
					p := h.clientsByID[res.PartnerClientID]
					h.mu.RUnlock()
					if p != nil {
						pair(c, p)
						h.Stats.Paired++
					}
				} else {
					// partner is remote → set remote link only
					c.Partner = &Client{ID: res.PartnerClientID, NodeID: res.PartnerNodeID}
					select {
					case c.Send <- []byte("🟢 Paired (remote)! Say hi."):
					default:
					}
					h.Stats.Paired++
				}

			}

		case c := <-h.Unregister:
			h.Stats.Connected--

			// Remove from local waiting in standalone
			for i, w := range h.Waiting {
				if w == c {
					copy(h.Waiting[i:], h.Waiting[i+1:])
					h.Waiting = h.Waiting[:len(h.Waiting)-1]
					h.Stats.Waiting--
					break
				}
			}

			// Remove from global queue if cluster
			if h.broker != nil {
				_ = h.broker.RemoveFromQueue(h.ctx, c.ID)
			}

			// If paired, notify the parner and unpair.
			if p := c.Partner; p != nil {
				unpair(c, p)
				// Offer partner back to waiting queue.
				h.Stats.Paired--
				h.Waiting = append(h.Waiting, p)
				h.Stats.Waiting++
			}
		}
	}
}

func (h *Hub) Snapshot() Stats {
	return *h.Stats
}
