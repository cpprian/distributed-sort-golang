package messages

import (
	"github.com/google/uuid"
)

type ConfirmMessage struct {
	BaseMessage
}

func NewConfirmMessage(transactionID uuid.UUID) ConfirmMessage {
	return ConfirmMessage{
		NewBaseMessageWithTransactionID(Confirm, transactionID),
	}
}
