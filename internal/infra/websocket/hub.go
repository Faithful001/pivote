package websocket

import (
	"encoding/json"
	"log"
)

type Hub struct {
	clients 	map[*Client] bool
	broadcast 	chan []byte
	incoming 	chan *IncomingMessage
	register 	chan *Client
	unregister	chan *Client
}

type IncomingMessage struct {
	Client	*Client
	Payload	[]byte
}

type IncomingEvent struct {
	Type	string `json:"type"`
}

func NewHub() *Hub {
	return &Hub {
		broadcast:  make(chan []byte),
		incoming:   make(chan *IncomingMessage, 64),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) Run() {
	for {
		select {
			case client := <- h.register:
				h.clients[client] = true
			case client := <- h.unregister:
				if _, ok := h.clients[client]; ok {
					delete(h.clients, client)
					close(client.send)
				}
			case message := <- h.broadcast:
				for client := range h.clients {
					select {
						case client.send <- message:
						default:
							delete(h.clients, client)
							close(client.send)
					}
				}

			case msg := <- h.incoming:
				h.RouteIncoming(msg)

		}
	}
}

func (h *Hub) RouteIncoming(msg *IncomingMessage) {
	var event IncomingEvent
	if err := json.Unmarshal(msg.Payload, &event); err != nil {
		log.Printf("ws: failed to parse incoming message: %v", err)
		return
	}

	switch event.Type {
	case "ping":
		h.HandlePing(msg.Client)
	default:
		log.Printf("ws: unknown event type %q - ignored", event.Type)
	}

}

func (h *Hub) HandlePing(client *Client) {
	pong, _ := json.Marshal(map[string]string{"type": "pong"})
	select {
	case client.send <- pong:
	default:
		close(client.send)
		delete(h.clients, client)
	}
}

func (h *Hub) HandleIncoming(client *Client, payload []byte) {
	h.incoming <- &IncomingMessage{Client: client, Payload: payload}
}

func (h *Hub) BroadcastToRaw(message []byte) {
	h.broadcast <- message
}

