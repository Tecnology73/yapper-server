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

func (h *Hub) broadcast(payload Payload, excludeClient *Client) {
	serializedPayload, err := json.Marshal(payload)
	if err != nil {
		return
	}

	for client := range h.clients {
		if client == excludeClient {
			continue
		}

		select {
		case client.send <- serializedPayload:
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
			if message.Payload.Type == Welcome {
				p := WelcomePacket{}
				err := json.Unmarshal(message.Payload.Data, &p)
				if err != nil {
					log.Printf("Error run#Unmarshal: %v", err)
					break
				}

				message.Client.user = store.GetOrCreateUser(p.Name)

				// Welcome response
				data, err := json.Marshal(WelcomeResponsePacket{
					Id: message.Client.user.Id,
				})
				if err != nil {
					log.Printf("Error run#Marshal#WelcomeResponsePacket: %v", err)
					break
				}

				message.Client.send <- data

				// User Connected broadcast
				data, err = json.Marshal(UserConnectedPacket{
					Id:   message.Client.user.Id,
					Name: p.Name,
				})
				if err != nil {
					log.Printf("Error run#Marshal#UserConnectedPacket: %v", err)
					break
				}

				h.broadcast(Payload{
					Type: UserConnected,
					Data: data,
				}, message.Client)

				break
			}

			if message.Client.user == nil {
				break
			}

			switch message.Payload.Type {
			case Welcome:
			case UserConnected:
			case UserDisconnected:
				break

			case ChatMessage:
				h.broadcast(message.Payload, message.Client)

			case StreamRoomsConnect:
			}
		}
	}
}
