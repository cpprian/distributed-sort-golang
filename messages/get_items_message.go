package messages

import (
	"github.com/google/uuid"
)

type GetItemsMessage struct {
	BaseMessage
	Items []int64 `json:"items"`
}

func NewGetItemsMessage() GetItemsMessage {
	return GetItemsMessage{
		BaseMessage: NewBaseMessage(GetItems),
		Items:       []int64{},
	}
}

func NewGetItemsMessageWithTransactionID(items []int64, transactionID uuid.UUID) GetItemsMessage {
	return GetItemsMessage{
		BaseMessage: NewBaseMessageWithTransactionID(GetItems, transactionID),
		Items:       items,
	}
}

func (m GetItemsMessage) GetItems() []int64 {
	return m.Items
}
