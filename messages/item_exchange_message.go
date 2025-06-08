package messages

import (
	"github.com/google/uuid"
)

type ItemExchangeMessage struct {
	BaseMessage
	OfferedItem int64 `json:"offeredItem"`
	WantedItem  int64 `json:"wantedItem"`
	SenderID    int64 `json:"senderId"`
}

func NewItemExchangeMessage(offered, wanted int64, senderID int64) ItemExchangeMessage {
	return ItemExchangeMessage{
		BaseMessage: NewBaseMessage(ItemExchange),
		OfferedItem: offered,
		WantedItem:  wanted,
		SenderID:    senderID,
	}
}

func NewItemExchangeMessageWithID(offered, wanted int64, transactionID uuid.UUID, senderID int64) ItemExchangeMessage {
	return ItemExchangeMessage{
		BaseMessage: NewBaseMessageWithTransactionID(ItemExchange, transactionID),
		OfferedItem: offered,
		WantedItem:  wanted,
		SenderID:    senderID,
	}
}

func (m ItemExchangeMessage) GetSenderID() int64 {
	return m.SenderID
}

func (m ItemExchangeMessage) GetOfferedItem() int64 {
	return m.OfferedItem
}

func (m ItemExchangeMessage) GetWantedItem() int64 {
	return m.WantedItem
}
