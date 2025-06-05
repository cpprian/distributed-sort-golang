package messages

import (
	"strconv"

	"github.com/cpprian/distributed-sort-golang/serializers"
	"github.com/google/uuid"
)

type AnnounceSelfMessage struct {
	Message
	ID               int64                     `json:"id"`
	ListeningAddress serializers.MultiaddrJSON `json:"listeningAddress"`
}

func NewAnnounceSelfMessage(id int64, addr serializers.MultiaddrJSON) AnnounceSelfMessage {
	return AnnounceSelfMessage{
		Message: Message{
			MessageType:   AnnounceSelf,
			TransactionID: uuid.New(),
		},
		ID:               id,
		ListeningAddress: addr,
	}
}

// String returns a string representation of the AnnounceSelfMessage.
func (m AnnounceSelfMessage) String() string {
	return "AnnounceSelfMessage{" +
		"ID: " + strconv.FormatInt(m.ID, 10) +
		", ListeningAddress: " + m.ListeningAddress.String() +
		"}"
}

func (m AnnounceSelfMessage) Type() MessageType {
	return m.MessageType
}
