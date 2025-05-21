package server

import (
	"net/http"

	"github.com/cloud-mill/webrtc-signalling/internal/logger"
	"github.com/cloud-mill/webrtc-signalling/internal/models"
	"github.com/olahol/melody"
	"go.uber.org/zap"
)

var (
	ClientPool *models.ClientPool
	m          = melody.New()
)

func init() {
	ClientPool = models.NewClientPool()
	go ClientPool.Start()

	registerConnectionHandlers()
	registerMessageHandlers()
	registerDisconnectHandlers()
}

func AcceptConnection(w http.ResponseWriter, r *http.Request) {
	if err := m.HandleRequest(w, r); err != nil {
		logger.Logger.Error("error accepting websocket connection", zap.Error(err))
		http.Error(w, "failed to establish websocket connection", http.StatusInternalServerError)
	}
}

func registerConnectionHandlers() {
	m.HandleConnect(func(s *melody.Session) {
		clientId := s.Request.URL.Query().Get("client_id")
		if clientId == "" {
			logger.Logger.Warn("missing client_id in connection request")
			_ = s.Close()
			return
		}

		client := models.NewClient(clientId, s, ClientPool)
		ClientPool.Register <- client

		s.Set("client_id", clientId)
		logger.Logger.Info("client connected", zap.String("client_id", clientId))
	})
}

func registerMessageHandlers() {
	m.HandleMessage(func(s *melody.Session, msg []byte) {
		clientIdVal, ok := s.Get("client_id")
		if !ok {
			logger.Logger.Warn("received message from unidentified session")
			return
		}
		clientId := clientIdVal.(string)

		client := ClientPool.GetClient(clientId)
		if client == nil {
			logger.Logger.Warn("client not found in pool", zap.String("client_id", clientId))
			return
		}

		if err := client.HandleMessage(msg); err != nil {
			logger.Logger.Error("failed to handle client message",
				zap.String("client_id", clientId),
				zap.Error(err),
			)
		}
	})
}

func registerDisconnectHandlers() {
	m.HandleDisconnect(func(s *melody.Session) {
		clientIdVal, ok := s.Get("client_id")
		if !ok {
			logger.Logger.Warn("session disconnected without client_id")
			return
		}
		clientId := clientIdVal.(string)

		client := ClientPool.GetClient(clientId)
		if client != nil {
			client.Leave()
		}

		logger.Logger.Info("client disconnected", zap.String("client_id", clientId))
	})
}
