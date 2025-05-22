package models

import (
	"encoding/json"

	"github.com/cloud-mill/webrtc-signalling/internal/logger"
	"github.com/olahol/melody"
	"go.uber.org/zap"
)

type Client struct {
	Id         string
	Session    *melody.Session
	ClientPool *ClientPool
	Rooms      map[string]bool
}

func NewClient(Id string, session *melody.Session, pool *ClientPool) *Client {
	logger.Logger.Info("new client connected", zap.String("client_Id", Id))
	return &Client{
		Id:         Id,
		Session:    session,
		ClientPool: pool,
		Rooms:      make(map[string]bool),
	}
}

func (c *Client) Write(message Message) {
	if c.Session == nil {
		logger.Logger.Warn("attempted to write to nil session", zap.String("client_Id", c.Id))
		return
	}

	data, err := json.Marshal(message)
	if err != nil {
		logger.Logger.Error("failed to marshal message",
			zap.String("client_Id", c.Id),
			zap.Error(err),
		)
		return
	}

	if err := c.Session.Write(data); err != nil {
		logger.Logger.Error("failed to send message to client",
			zap.String("client_Id", c.Id),
			zap.Error(err),
		)
	}
}

func (c *Client) WriteRaw(data []byte) {
	if c.Session != nil {
		_ = c.Session.Write(data)
	}
}

func (c *Client) HandleMessage(raw []byte) error {
	logger.Logger.Info("received message from client",
		zap.String("client_Id", c.Id),
		zap.ByteString("data", raw),
	)

	var msg Message
	if err := json.Unmarshal(raw, &msg); err != nil {
		logger.Logger.Error("invalid message format",
			zap.String("client_Id", c.Id),
			zap.Error(err),
		)
		return err
	}

	return c.routeMessage(msg)
}

func (c *Client) routeMessage(msg Message) error {
	switch msg.MessageType {
	case ClientJoinRoom:
		return c.JoinRoom(msg.RoomId)

	case ClientLeaveRoom:
		return c.LeaveRoom(msg.RoomId)

	case Broadcast:
		return c.BroadcastToRoom(msg.RoomId, msg)

	case Offer, Answer, Candidate:
		return c.ForwardToPeer(msg)

	default:
		logger.Logger.Warn("unhandled message type",
			zap.String("client_Id", c.Id),
			zap.String("type", string(msg.MessageType)),
		)
		return nil
	}
}

func (c *Client) JoinRoom(roomId string) error {
	if roomId == "" {
		logger.Logger.Warn("empty room Id", zap.String("client_Id", c.Id))
		return nil
	}
	JoinRoom(c, roomId)
	c.Rooms[roomId] = true

	logger.Logger.Info("joined room",
		zap.String("client_Id", c.Id),
		zap.String("room_Id", roomId),
	)
	return nil
}

func (c *Client) LeaveRoom(roomId string) error {
	if roomId == "" {
		return nil
	}
	LeaveRoom(c, roomId)
	delete(c.Rooms, roomId)

	logger.Logger.Info("left room",
		zap.String("client_Id", c.Id),
		zap.String("room_Id", roomId),
	)
	return nil
}

func (c *Client) BroadcastToRoom(roomId string, msg Message) error {
	BroadcastToRoom(c, roomId, msg)
	return nil
}

func (c *Client) ForwardToPeer(msg Message) error {
	peer := c.ClientPool.GetClient(msg.To)
	if peer == nil {
		logger.Logger.Warn("target peer not found",
			zap.String("client_Id", c.Id),
			zap.String("target_Id", msg.To),
		)
		return nil
	}
	peer.Write(msg)
	return nil
}

func (c *Client) Leave() {
	if c.ClientPool != nil {
		for roomId := range c.Rooms {
			if err := c.LeaveRoom(roomId); err != nil {
				logger.Logger.Warn("failed to leave room",
					zap.String("client_Id", c.Id),
					zap.String("room_Id", roomId),
					zap.Error(err),
				)
			}
		}
		c.ClientPool.Unregister <- c
	}
	logger.Logger.Info("client disconnected", zap.String("client_Id", c.Id))
}
