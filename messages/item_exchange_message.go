package messages

import (
	"github.com/google/uuid"
)

type ItemExchangeMessage struct {
	Message
	OfferedItem int   `json:"offeredItem"`
	WantedItem  int   `json:"wantedItem"`
	SenderID    int64 `json:"senderId"`
	Response    bool  `json:"response,omitempty"`
}

func NewItemExchangeMessage(offered, wanted int, senderID int64) ItemExchangeMessage {
	return ItemExchangeMessage{
		Message:     NewMessage(ItemExchange),
		OfferedItem: offered,
		WantedItem:  wanted,
		SenderID:    senderID,
	}
}

func NewItemExchangeMessageWithID(offered, wanted int, id uuid.UUID, senderID int64) ItemExchangeMessage {
	return ItemExchangeMessage{
		Message:     Message{MessageType: ItemExchange, TransactionID: id},
		OfferedItem: offered,
		WantedItem:  wanted,
		SenderID:    senderID,
	}
}
