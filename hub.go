package main

import (
	"encoding/json"
	"log"
)

type Hub struct {
	clients    map[*Client]bool
	message    chan Message
	register   chan *Client
	unregister chan *Client
}

var store = &Store{
	Users: make(map[string]*StoreUser),
	Rooms: make(map[string]*StoreRoom),
}

func newHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		message:    make(chan Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) send(msgType MessageType, data []byte, client *Client) {
	payload := make([]byte, 1+len(data))
	payload[0] = uint8(msgType)
	copy(payload[1:], data)

	client.send <- payload
}

func (h *Hub) broadcast(msgType MessageType, data []byte, excludeClient *Client) {
	// If there are no clients, or the only client connected is who we want to exclude.
	if len(h.clients) == 0 || (len(h.clients) == 1 && h.clients[excludeClient]) {
		return
	}

	payload := make([]byte, 1+len(data))
	payload[0] = uint8(msgType)
	copy(payload[1:], data)

	for client := range h.clients {
		if client == excludeClient {
			continue
		}

		select {
		case client.send <- payload:
		default:
			delete(h.clients, client)
			close(client.send)
		}
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}

		case message := <-h.message:
			if message.Type == Welcome {
				p := WelcomePacket{}
				err := json.Unmarshal(message.Data, &p)
				if err != nil {
					log.Printf("Error run#Unmarshal: %v", err)
					break
				}

				message.Client.user = store.GetOrCreateUser(p.Name)

				// Welcome response
				users := make(map[uint64]string, len(h.clients))
				for client := range h.clients {
					users[client.user.Id] = client.user.Name
				}

				payload, err := json.Marshal(WelcomeResponsePacket{
					Id:    message.Client.user.Id,
					Users: users,
				})
				if err != nil {
					break
				}

				h.send(Welcome, payload, message.Client)

				// User Connected broadcast
				payload, err = json.Marshal(UserConnectedPacket{
					Id:   message.Client.user.Id,
					Name: p.Name,
				})
				if err != nil {
					break
				}

				h.broadcast(UserConnected, payload, message.Client)

				break
			}

			if message.Client.user == nil {
				break
			}

			switch message.Type {
			case Welcome:
			case UserConnected:
			case UserDisconnected:
				h.broadcast(message.Type, message.Data, message.Client)
				break

			case ChatMessage:
				h.broadcast(message.Type, message.Data, message.Client)
				break

			case StreamRoomsConnect:
			}
		}
	}
}
