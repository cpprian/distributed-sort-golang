package messages

import (
	"github.com/google/uuid"
)

type GetItemsMessage[T any] struct {
	Message
	Items []T `json:"items"`
}

func NewGetItemsMessage[T any]() GetItemsMessage[T] {
	return GetItemsMessage[T]{
		Message: NewMessage(GetItems),
	}
}

func NewGetItemsMessageWithID[T any](items []T, txID uuid.UUID) GetItemsMessage[T] {
	return GetItemsMessage[T]{
		Message: Message{
			MessageType:   GetItems,
			TransactionID: txID,
		},
		Items: items,
	}
}
