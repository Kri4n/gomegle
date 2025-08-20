package server

type Hub struct {
	Register   chan *Client
	Unregister chan *Client
	Waiting    []*Client
	Stats      *Stats
}

type Stats struct {
	Connected int
	Waiting   int
	Paired    int
}

func NewHub() *Hub {
	return &Hub{
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Waiting:    make([]*Client, 0, 64),
		Stats:      &Stats{},
	}
}

func (h *Hub) Run() {
	for {
		select {
		case c := <-h.Register:
			h.Stats.Connected++
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

		case c := <-h.Unregister:
			h.Stats.Connected--
			// Remove from waiting if present.
			for i, w := range h.Waiting {
				if w == c {
					// delete index i
					copy(h.Waiting[i:], h.Waiting[i+1:])
					h.Waiting = h.Waiting[:len(h.Waiting)-1]
					h.Stats.Waiting--
					break
				}
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