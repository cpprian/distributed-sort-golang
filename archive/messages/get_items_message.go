package messages

import (
	"github.com/google/uuid"
)

type GetItemsMessage struct {
	Message
	Items    []int64 `json:"items"`
	SenderID int64   `json:"senderId,omitempty"`
}

func NewGetItemsMessage() GetItemsMessage {
	return GetItemsMessage{
		Message: NewMessage(GetItems),
	}
}

func NewGetItemsMessageWithID(items []int64, txID uuid.UUID) GetItemsMessage {
	return GetItemsMessage{
		Message: Message{
			MessageType:   GetItems,
			TransactionID: txID,
		},
		Items: items,
	}
}
