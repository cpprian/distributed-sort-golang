package messages

import (
	"github.com/google/uuid"
)

type ItemExchangeMessage struct {
	Message
	OfferedItem int64   `json:"offeredItem"`
	WantedItem  int64   `json:"wantedItem"`
	SenderID    int64 `json:"senderId"`
	Response    bool  `json:"response,omitempty"`
}

func NewItemExchangeMessage(offered, wanted int64, senderID int64) ItemExchangeMessage {
	return ItemExchangeMessage{
		Message:     NewMessage(ItemExchange),
		OfferedItem: offered,
		WantedItem:  wanted,
		SenderID:    senderID,
	}
}

func NewItemExchangeMessageWithID(offered, wanted int64, id uuid.UUID, senderID int64) ItemExchangeMessage {
	return ItemExchangeMessage{
		Message:     Message{MessageType: ItemExchange, TransactionID: id},
		OfferedItem: offered,
		WantedItem:  wanted,
		SenderID:    senderID,
	}
}
