package messages

import (
	"github.com/google/uuid"
)

type GetItemsMessage struct {
	Message
	Items    []int `json:"items"`
	SenderID int64 `json:"senderId,omitempty"`
}

func NewGetItemsMessage() GetItemsMessage {
	return GetItemsMessage{
		Message: NewMessage(GetItems),
	}
}

func NewGetItemsMessageWithID(items []int, txID uuid.UUID) GetItemsMessage {
	return GetItemsMessage{
		Message: Message{
			MessageType:   GetItems,
			TransactionID: txID,
		},
		Items: items,
	}
}
