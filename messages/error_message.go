package messages

import "github.com/google/uuid"

type ErrorMessage struct {
	Message
}

func NewErrorMessage(txID uuid.UUID) ErrorMessage {
	return ErrorMessage{
		Message: Message{
			MessageType:   ErrorType,
			TransactionID: txID,
		},
	}
}

func (m ErrorMessage) String() string {
	return "ErrorMessage{" +
		"TransactionID: " + m.TransactionID.String() +
		"}"
}

func (m ErrorMessage) Type() MessageType {
	return m.MessageType
}
