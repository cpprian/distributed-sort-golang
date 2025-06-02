package messages

import (
	"github.com/google/uuid"
	"reflect"
)

// MessageType represents the type of a message.
type MessageType string

const (
	ItemExchange         MessageType = "ITEM_EXCHANGE"
	CornerItemChange     MessageType = "CORNER_ITEM_CHANGE"
	NodesList            MessageType = "NODES_LIST"
	NodesListResponse    MessageType = "NODES_LIST_RESPONSE"
	AnnounceSelf         MessageType = "ANNOUNCE_SELF"
	GetItems             MessageType = "GET_ITEMS"
	Confirm              MessageType = "CONFIRM"
	ErrorType            MessageType = "ERROR"
)

// TypeMetadata holds metadata for each MessageType.
type TypeMetadata struct {
	GoType         reflect.Type
	RequireResponse bool
}

// MessageRegistry maps MessageTypes to their metadata.
var MessageRegistry = map[MessageType]TypeMetadata{
	ItemExchange:      {GoType: reflect.TypeOf(ItemExchangeMessage{}), RequireResponse: true},
	CornerItemChange:  {GoType: reflect.TypeOf(CornerItemChangeMessage[any]{}), RequireResponse: true},
	NodesList:         {GoType: reflect.TypeOf(NodesListMessage{}), RequireResponse: true},
	NodesListResponse: {GoType: reflect.TypeOf(NodesListResponseMessage{}), RequireResponse: false},
	AnnounceSelf:      {GoType: reflect.TypeOf(AnnounceSelfMessage{}), RequireResponse: false},
	GetItems:          {GoType: reflect.TypeOf(GetItemsMessage{}), RequireResponse: true},
	Confirm:           {GoType: reflect.TypeOf(ConfirmMessage{}), RequireResponse: false},
	ErrorType:         {GoType: reflect.TypeOf(ErrorMessage{}), RequireResponse: false},
}

// Message is the base struct embedded in all messages.
type Message struct {
	MessageType   MessageType  `json:"messageType"`
	TransactionID uuid.UUID    `json:"transactionId"`
}

// NewMessage initializes a base Message.
func NewMessage(msgType MessageType) Message {
	return Message{
		MessageType:   msgType,
		TransactionID: uuid.New(),
	}
}
