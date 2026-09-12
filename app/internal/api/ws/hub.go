package ws

type DirectMessage struct {
	SenderID    int64
	RecipientID int64
	ConvID      int64
	Payload     []byte
}

type Hub struct {
	// support multiple client per user
	users map[int64]map[*Client]bool

	// set active clients in conversation
	convs map[int64]map[*Client]bool

	SendDirect chan *DirectMessage

	Register   chan *Client
	UnRegister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		users:      make(map[int64]map[*Client]bool),
		convs:      make(map[int64]map[*Client]bool),
		SendDirect: make(chan *DirectMessage, 256),
		Register:   make(chan *Client, 32),
		UnRegister: make(chan *Client, 32),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			// Register client to users map
			if _, ok := h.users[client.UserID]; !ok {
				h.users[client.UserID] = make(map[*Client]bool)
			}
			h.users[client.UserID][client] = true

			// Register client to convs map if convID is set
			if client.ConvID > 0 {
				if _, ok := h.users[client.ConvID]; !ok {
					h.users[client.ConvID] = make(map[*Client]bool)
				}
				h.users[client.ConvID][client] = true
			}

		case client := <-h.UnRegister:
			// Unregister clent from users map
			if clients, ok := h.users[client.UserID]; ok {
				if _, exists := clients[client]; exists {
					delete(clients, client)
					close(client.Send)
					if len(clients) == 0 {
						delete(h.users, client.UserID)
					}
				}
			}

			// Unregister client from convs map
			if client.ConvID > 0 {
				if clients, ok := h.convs[client.ConvID]; ok {
					if _, exists := clients[client]; exists {
						delete(clients, client)
						if len(clients) == 0 {
							delete(h.convs, client.ConvID)
						}
					}
				}
			}

		case msg := <-h.SendDirect:
			// Deliver direct message to recipient
			if msg.RecipientID > 0 {
				if recipientClients, ok := h.users[msg.RecipientID]; ok {
					for c := range recipientClients {
						select {
						case c.Send <- msg.Payload:
						default:
							h.dropClient(c)
						}
					}

					// Echo back to the sender's client(s) for delivery confirmation
					if senderClints, ok := h.users[msg.SenderID]; ok {
						for c := range senderClints {
							select {
							case c.Send <- msg.Payload:
							default:
								h.dropClient(c)

							}
						}
					}
					continue
				}
			}
		}
	}
}

func (h *Hub) dropClient(c *Client) {
	if clients, ok := h.users[c.UserID]; ok {
		if _, exists := clients[c]; exists {
			delete(clients, c)
			if len(clients) == 0 {
				delete(h.users, c.UserID)
			}
		}
	}

	if c.ConvID > 0 {
		if clients, ok := h.convs[c.ConvID]; ok {
			delete(clients, c)
			if len(clients) == 0 {
				delete(h.convs, c.ConvID)
			}
		}
	}

	close(c.Send)
}
