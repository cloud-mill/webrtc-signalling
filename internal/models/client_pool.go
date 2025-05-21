package models

import (
	"sync"

	"github.com/cloud-mill/webrtc-signalling/internal/logger"
	"go.uber.org/zap"
)

type ClientPool struct {
	Register   chan *Client
	Unregister chan *Client
	Clients    map[string]*Client
	mu         sync.RWMutex
}

func NewClientPool() *ClientPool {
	return &ClientPool{
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Clients:    make(map[string]*Client),
	}
}

func (cp *ClientPool) Start() {
	for {
		select {
		case client := <-cp.Register:
			cp.handleRegister(client)

		case client := <-cp.Unregister:
			cp.handleUnregister(client)
		}
	}
}

func (cp *ClientPool) handleRegister(client *Client) {
	if client == nil {
		return
	}
	logger.Logger.Info("registering client", zap.String("client_id", client.Id))
	cp.SetClient(client.Id, client)
}

func (cp *ClientPool) handleUnregister(client *Client) {
	if client == nil {
		return
	}
	logger.Logger.Info("unregistering client", zap.String("client_id", client.Id))
	cp.DeleteClient(client.Id)

	if client.Session != nil {
		if err := client.Session.Close(); err != nil {
			logger.Logger.Error("failed to close client session",
				zap.String("client_id", client.Id),
				zap.Error(err),
			)
		}
	}
}

func (cp *ClientPool) GetClient(clientID string) *Client {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	return cp.Clients[clientID]
}

func (cp *ClientPool) SetClient(clientID string, client *Client) {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	cp.Clients[clientID] = client
}

func (cp *ClientPool) DeleteClient(clientID string) {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	delete(cp.Clients, clientID)
}

func (cp *ClientPool) SendMessageToClient(clientID string, message Message) {
	client := cp.GetClient(clientID)
	if client != nil {
		client.Write(message)
	}
}

func (cp *ClientPool) ClientExitFromPool(clientID string) {
	client := cp.GetClient(clientID)
	if client != nil {
		cp.Unregister <- client
	}
}
