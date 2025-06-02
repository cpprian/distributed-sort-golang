package messages

import (
	"github.com/cpprian/distributed-sort-golang/serializers"
)

type AnnounceSelfMessage struct {
	Message
	ID               int64                     `json:"id"`
	ListeningAddress serializers.MultiaddrJSON `json:"listeningAddress"`
}

func NewAnnounceSelfMessage(id int64, addr serializers.MultiaddrJSON) AnnounceSelfMessage {
	return AnnounceSelfMessage{
		Message: Message{
			Type: MessageAnnounceSelf,
		},
		ID:               id,
		ListeningAddress: addr,
	}
}
