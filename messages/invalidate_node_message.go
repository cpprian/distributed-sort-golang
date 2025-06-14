package messages

import "github.com/google/uuid"

type InvalidateNodeMessage struct {
	BaseMessage
	ID int64 `json:"id"`
}

func NewInvalidateNodeMessage(id int64) InvalidateNodeMessage {
	return InvalidateNodeMessage{
		BaseMessage: NewBaseMessage(InvalidateNode),
		ID:          id,
	}
}

func NewInvalidateNodeMessageWithTransactionID(id int64, transactionID uuid.UUID) InvalidateNodeMessage {
	return InvalidateNodeMessage{
		BaseMessage: NewBaseMessageWithTransactionID(InvalidateNode, transactionID),
		ID:          id,
	}
}

func (m InvalidateNodeMessage) GetMessageType() MessageType {
	return InvalidateNode
}

func (m InvalidateNodeMessage) GetTransactionID() uuid.UUID {
	return m.TransactionID
}

func (m InvalidateNodeMessage) GetID() int64 {
	return m.ID
}
