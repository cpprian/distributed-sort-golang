package messages

import "github.com/google/uuid"

type ItemsBackupMessage struct {
	BaseMessage
	Items    []int64 `json:"items"`
	SenderID int64   `json:"senderId"`
}

func NewItemsBackupMessage(items []int64, senderID int64) *ItemsBackupMessage {
	return &ItemsBackupMessage{
		BaseMessage: NewBaseMessage(ItemsBackup),
		Items:       items,
		SenderID:    senderID,
	}
}

func NewItemsBackupMessageWithTransactionID(items []int64, senderID int64, transactionID uuid.UUID) *ItemsBackupMessage {
	return &ItemsBackupMessage{
		BaseMessage: NewBaseMessageWithTransactionID(ItemsBackup, transactionID),
		Items:       items,
		SenderID:    senderID,
	}
}

func (m *ItemsBackupMessage) GetMessageType() MessageType {
	return ItemsBackup
}

func (m *ItemsBackupMessage) GetTransactionID() uuid.UUID {
	return m.TransactionID
}

func (m *ItemsBackupMessage) GetItems() []int64 {
	return m.Items
}

func (m *ItemsBackupMessage) GetSenderID() int64 {
	return m.SenderID
}
