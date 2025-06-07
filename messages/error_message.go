package messages

import (
	"github.com/google/uuid"
)

type ErrorMessage struct {
	BaseMessage
}

func NewErrorMessage(transactionID uuid.UUID) ErrorMessage {
	return ErrorMessage{
		NewBaseMessageWithTransactionID(Error, transactionID),
	}
}
