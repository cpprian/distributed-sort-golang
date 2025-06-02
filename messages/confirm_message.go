package messages

import "github.com/google/uuid"

type ConfirmMessage struct {
	Message
}

func NewConfirmMessage(txID uuid.UUID) ConfirmMessage {
	return ConfirmMessage{
		Message: Message{
			MessageType:   Confirm,
			TransactionID: txID,
		},
	}
}
