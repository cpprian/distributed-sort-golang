package messages

import (
	ma "github.com/multiformats/go-multiaddr"
)

type AnnounceSelfMessage struct {
	BaseMessage
	ID               int64        `json:"id"`
	ListeningAddress ma.Multiaddr `json:"listeningAddress"`
}

func NewAnnounceSelfMessage(id int64, listeningAddress ma.Multiaddr) AnnounceSelfMessage {
	return AnnounceSelfMessage{
		BaseMessage:      NewBaseMessage(AnnounceSelf),
		ID:               id,
		ListeningAddress: listeningAddress,
	}
}
