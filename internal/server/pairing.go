package server

func pair(a, b *Client) {
	a.Partner = b
	b.Partner = a
	// Soft notifications
	select {
	case a.Send <- []byte("Paired! Say Hi."):
	default:
	}
	select {
	case b.Send <- []byte("Paired! Say Hi."):
	default:
	}
}

func unpair(a, b *Client) {
	if a.Partner == b {
		a.Partner = nil
	}
	if b.Partner == a {
		b.Partner = nil
	}
	select {
	case b.Send <- []byte("Your partner has disconnected. Searching..."):
	default:
	}
}