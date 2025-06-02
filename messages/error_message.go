package messages

import "github.com/google/uuid"

type ErrorMessage struct {
	Message
}

func NewErrorMessage(txID uuid.UUID) ErrorMessage {
	return ErrorMessage{
		Message: Message{
			Type:          MessageError,
			TransactionID: &txID,
		},
	}
}
