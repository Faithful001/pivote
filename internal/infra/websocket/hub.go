package websocket

import (
	"encoding/json"
	"log"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Incoming messages from individual clients to be routed.
	incoming chan *IncomingMessage

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

// IncomingMessage wraps a raw payload with the originating client so handlers
// can send a reply back to that specific client only.
type IncomingMessage struct {
	Client  *Client
	Payload []byte
}

// incomingEvent is used to decode just the "type" field of any client message.
type incomingEvent struct {
	Type string `json:"type"`
}

func NewHub() *Hub {
	return &Hub{
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
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		case msg := <-h.incoming:
			h.routeIncoming(msg)
		}
	}
}

// routeIncoming dispatches a client message to the appropriate handler based
// on its "type" field.
func (h *Hub) routeIncoming(msg *IncomingMessage) {
	var event incomingEvent
	if err := json.Unmarshal(msg.Payload, &event); err != nil {
		log.Printf("ws: failed to parse incoming message: %v", err)
		return
	}

	switch event.Type {
	case "ping":
		h.handlePing(msg.Client)
	default:
		log.Printf("ws: unknown event type %q - ignored", event.Type)
	}
}

// handlePing replies with a pong directly to the sender.
func (h *Hub) handlePing(client *Client) {
	pong, _ := json.Marshal(map[string]string{"type": "pong"})
	select {
	case client.send <- pong:
	default:
		close(client.send)
		delete(h.clients, client)
	}
}

// HandleIncoming is called by the client's readPump to forward a received
// message into the hub for routing.
func (h *Hub) HandleIncoming(client *Client, payload []byte) {
	h.incoming <- &IncomingMessage{Client: client, Payload: payload}
}

// BroadcastToRaw sends a raw byte message to all connected clients.
func (h *Hub) BroadcastToRaw(message []byte) {
	h.broadcast <- message
}
