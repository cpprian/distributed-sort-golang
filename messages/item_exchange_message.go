package messages

import (
	"github.com/google/uuid"
)

type ItemExchangeMessage[T any] struct {
	Message
	OfferedItem T    `json:"offeredItem"`
	WantedItem  T    `json:"wantedItem"`
	SenderID    int64 `json:"senderId"`
}

func NewItemExchangeMessage[T any](offered, wanted T, senderID int64) ItemExchangeMessage[T] {
	return ItemExchangeMessage[T]{
		Message:     NewMessage(ItemExchange),
		OfferedItem: offered,
		WantedItem:  wanted,
		SenderID:    senderID,
	}
}

func NewItemExchangeMessageWithID[T any](offered, wanted T, id uuid.UUID, senderID int64) ItemExchangeMessage[T] {
	return ItemExchangeMessage[T]{
		Message:     Message{MessageType: ItemExchange, TransactionID: id},
		OfferedItem: offered,
		WantedItem:  wanted,
		SenderID:    senderID,
	}
}
