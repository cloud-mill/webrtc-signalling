package models

import "time"

type MessageType string

const (
	Broadcaster    MessageType = "broadcaster"
	Connect        MessageType = "connect"
	Watcher        MessageType = "watcher"
	Offer          MessageType = "offer"
	Answer         MessageType = "answer"
	Candidate      MessageType = "candidate"
	Disconnect     MessageType = "disconnect"
	DisconnectPeer MessageType = "disconnectPeer"

	ClientJoinRoom  MessageType = "joinRoom"
	ClientLeaveRoom MessageType = "leaveRoom"
	Broadcast       MessageType = "broadcast"

	Ping MessageType = "ping"
	Pong MessageType = "pong"

	Error MessageType = "error"
	Ack   MessageType = "ack"
)

type Message struct {
	From           string      `json:"from"`
	To             string      `json:"to,omitempty"`
	RoomId         string      `json:"roomId,omitempty"`
	MessageType    MessageType `json:"messageType"`
	MessageContent interface{} `json:"messageContent,omitempty"`
	Timestamp      time.Time   `json:"timestamp,omitempty"`
}
