package messages

import (
	"strconv"

	"github.com/google/uuid"
)

type CornerItemChangeMessage struct {
	Message
	Item      int64  `json:"item"`
	SenderID  int64  `json:"senderId"`
	Direction string `json:"direction,omitempty"` // Optional field for direction, if needed
}

func NewCornerItemChangeMessage(item int64, senderID int64, direction string) CornerItemChangeMessage {
	return CornerItemChangeMessage{
		Message: Message{
			MessageType:   CornerItemChange,
			TransactionID: uuid.New(),
		},
		Item:      item,
		SenderID:  senderID,
		Direction: direction,
	}
}

func (m CornerItemChangeMessage) String() string {
	return "CornerItemChangeMessage{" +
		"Item: " + strconv.FormatInt(m.Item, 10) +
		", SenderID: " + strconv.FormatInt(m.SenderID, 10) +
		", Direction: " + m.Direction +
		"}"
}

func (m CornerItemChangeMessage) Type() MessageType {
	return m.MessageType
}
