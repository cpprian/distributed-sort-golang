package messages

import (
	"reflect"
	"github.com/google/uuid"
)

type MessageType int

const (
	ItemExchange MessageType = iota
	CornerItemChange
	NodesList
	NodesListResponse
	AnnounceSelf
	GetItems
	Confirm
	Error
)

type MessageInfo struct {
	GoType           reflect.Type
	RequiresResponse bool
}

var messageRegistry = map[MessageType]MessageInfo{
	ItemExchange:      {reflect.TypeOf(ItemExchangeMessage{}), true},
	CornerItemChange:  {reflect.TypeOf(CornerItemChangeMessage{}), true},
	NodesList:         {reflect.TypeOf(NodesListMessage{}), true},
	NodesListResponse: {reflect.TypeOf(NodesListResponseMessage{}), false},
	AnnounceSelf:      {reflect.TypeOf(AnnounceSelfMessage{}), false},
	GetItems:          {reflect.TypeOf(GetItemsMessage{}), true},
	Confirm:           {reflect.TypeOf(ConfirmMessage{}), false},
	Error:             {reflect.TypeOf(ErrorMessage{}), false},
}

type IMessage interface {
	GetMessageType() MessageType
	GetTransactionID() uuid.UUID
}

type BaseMessage struct {
	MessageType   MessageType `json:"messageType"`
	TransactionID uuid.UUID   `json:"transactionId"`
}

func NewBaseMessage(msgType MessageType) BaseMessage {
	return BaseMessage{
		MessageType:   msgType,
		TransactionID: uuid.New(),
	}
}

func NewBaseMessageWithTransactionID(msgType MessageType, transactionID uuid.UUID) BaseMessage {
	return BaseMessage{
		MessageType:   msgType,
		TransactionID: transactionID,
	}
}

func (m BaseMessage) GetMessageType() MessageType {
	return m.MessageType
}

func (m BaseMessage) GetTransactionID() uuid.UUID {
	return m.TransactionID
}
