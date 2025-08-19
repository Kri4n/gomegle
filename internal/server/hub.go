package server

type Hub struct {
	Register   chan *Client
	Unregister chan *Client
	Waiting    []*Client
}

func NewHub() *Hub {
	return &Hub{
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Waiting:    make([]*Client, 0, 64),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case c := <-h.Register:
			// Try to pair immediately.
			if len(h.Waiting) > 0 {
				partner := h.Waiting[0]
				h.Waiting = h.Waiting[1:]
				pair(c, partner)
			} else {
				h.Waiting = append(h.Waiting, c)
			}

		case c := <-h.Unregister:
			// Remove from waiting if present.
			for i, w := range h.Waiting {
				if w == c {
					// delete index i
					copy(h.Waiting[i:], h.Waiting[i+1:])
					h.Waiting = h.Waiting[:len(h.Waiting)-1]
					break
				}
			}
			// If paired, notify the parnet and unpair.
			if p := c.Partner; p != nil {
				unpair(c, p)
				// Offer partner back to waiting queue.
				h.Waiting = append(h.Waiting, p)
			}
			close(c.Send)
		}
	}
}