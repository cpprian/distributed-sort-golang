package messages

import "github.com/google/uuid"

type ConfirmMessage struct {
	Message
	SenderID      int64     `json:"senderId,omitempty"`
}

func NewConfirmMessage(txID uuid.UUID, senderID int64) ConfirmMessage {
	return ConfirmMessage{
		Message: Message{
			MessageType:   Confirm,
			TransactionID: txID,
		},
		SenderID:      senderID,
	}
}

func (m ConfirmMessage) String() string {
	return "ConfirmMessage{" +
		"TransactionID: " + m.TransactionID.String() +
		", SenderID: " + string(m.SenderID) +
		"}"
}

func (m ConfirmMessage) Type() MessageType {
	return m.MessageType
}
