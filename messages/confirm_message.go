package messages

import "github.com/google/uuid"

type ConfirmMessage struct {
	Message
	TransactionID uuid.UUID `json:"transactionId"`
	SenderID      int64     `json:"senderId,omitempty"`
}

func NewConfirmMessage(txID uuid.UUID, senderID int64) ConfirmMessage {
	return ConfirmMessage{
		Message: Message{
			MessageType:   Confirm,
			TransactionID: txID,
		},
		TransactionID: txID,
		SenderID:      senderID,
	}
}
