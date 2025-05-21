package models

import (
	"sync"
)

type Room struct {
	Id      string
	Members map[string]*Client
	mu      sync.RWMutex
}

var (
	Rooms   = make(map[string]*Room)
	roomsMu sync.RWMutex
)

func JoinRoom(client *Client, roomId string) {
	if client == nil {
		return
	}

	roomsMu.Lock()
	room, exists := Rooms[roomId]
	if !exists {
		room = &Room{
			Id:      roomId,
			Members: make(map[string]*Client),
		}
		Rooms[roomId] = room
	}
	roomsMu.Unlock()

	room.mu.Lock()
	room.Members[client.Id] = client
	room.mu.Unlock()

	if client.Rooms == nil {
		client.Rooms = make(map[string]bool)
	}
	client.Rooms[roomId] = true
}

func LeaveRoom(client *Client, roomId string) {
	if client == nil {
		return
	}

	roomsMu.RLock()
	room, ok := Rooms[roomId]
	roomsMu.RUnlock()
	if !ok {
		return
	}

	room.mu.Lock()
	delete(room.Members, client.Id)
	room.mu.Unlock()

	delete(client.Rooms, roomId)
}

func BroadcastToRoom(sender *Client, roomId string, msg Message) {
	if sender == nil {
		return
	}

	roomsMu.RLock()
	room, ok := Rooms[roomId]
	roomsMu.RUnlock()

	if !ok || !sender.Rooms[roomId] {
		return
	}

	room.mu.RLock()
	defer room.mu.RUnlock()

	for id, member := range room.Members {
		if id != sender.Id {
			member.Write(msg)
		}
	}
}
