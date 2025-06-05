package messages

import (
	"fmt"
	"log"
	"reflect"
	"strconv"

	"github.com/cpprian/distributed-sort-golang/neighbour"
	"github.com/cpprian/distributed-sort-golang/serializers"
	"github.com/google/uuid"
)

// MessageType represents the type of a message.
type MessageType string

const (
	ItemExchange      MessageType = "ITEM_EXCHANGE"
	CornerItemChange  MessageType = "CORNER_ITEM_CHANGE"
	NodesList         MessageType = "NODES_LIST"
	NodesListResponse MessageType = "NODES_LIST_RESPONSE"
	AnnounceSelf      MessageType = "ANNOUNCE_SELF"
	GetItems          MessageType = "GET_ITEMS"
	Confirm           MessageType = "CONFIRM"
	ErrorType         MessageType = "ERROR"
)

// TypeMetadata holds metadata for each MessageType.
type TypeMetadata struct {
	GoType          reflect.Type
	RequireResponse bool
}

// MessageRegistry maps MessageTypes to their metadata.
var MessageRegistry = map[MessageType]TypeMetadata{
	ItemExchange:      {GoType: reflect.TypeOf(ItemExchangeMessage{}), RequireResponse: true},
	CornerItemChange:  {GoType: reflect.TypeOf(CornerItemChangeMessage{}), RequireResponse: true},
	NodesList:         {GoType: reflect.TypeOf(NodesListMessage{}), RequireResponse: true},
	NodesListResponse: {GoType: reflect.TypeOf(NodesListResponseMessage{}), RequireResponse: false},
	AnnounceSelf:      {GoType: reflect.TypeOf(AnnounceSelfMessage{}), RequireResponse: false},
	GetItems:          {GoType: reflect.TypeOf(GetItemsMessage{}), RequireResponse: true},
	Confirm:           {GoType: reflect.TypeOf(ConfirmMessage{}), RequireResponse: false},
	ErrorType:         {GoType: reflect.TypeOf(ErrorMessage{}), RequireResponse: false},
}

// Message is the base interface for all messages in the system.
type Message struct {
	MessageType   MessageType `json:"messageType"`
	TransactionID uuid.UUID   `json:"transactionId"`
}

// MessageInterface is the interface that all message types must implement.
type MessageInterface interface {
	// Type returns the MessageType of the message.
	Type() MessageType
}

// NewMessage initializes a base Message.
func NewMessage(msgType MessageType) Message {
	return Message{
		MessageType:   msgType,
		TransactionID: uuid.New(),
	}
}

// Type returns the MessageType of the message.
func (m Message) Type() MessageType {
	return m.MessageType
}

// String returns a string representation of the Message.
func (m Message) String() string {
	return "Message{" +
		"MessageType: " + string(m.MessageType) +
		", TransactionID: " + m.TransactionID.String() +
		"}"
}

// DeserializeMessage deserializes a map into a MessageInterface.
func DeserializeMessage(data map[string]interface{}, msgType MessageType) (MessageInterface, error) {
	if data == nil {
		return nil, fmt.Errorf("data cannot be nil")
	}
	var msg MessageInterface

	tranID, err := uuid.Parse(data["transactionId"].(string))
	if err != nil {
		log.Println("Failed to parse transaction ID:", err)
		return nil, fmt.Errorf("failed to parse transaction ID: %w", err)
	}

	switch msgType {
	case CornerItemChange:
		msg = CornerItemChangeMessage{
			Message: Message{
				MessageType:   msgType,
				TransactionID: tranID,
			},
			SenderID:  int64(data["senderId"].(float64)),
			Item:      int64(data["item"].(float64)),
			Direction: data["direction"].(string),
		}
	case ItemExchange:
		tranID, err := uuid.Parse(data["transactionId"].(string))
		if err != nil {
			log.Println("Failed to parse transaction ID:", err)
			return nil, fmt.Errorf("failed to parse transaction ID: %w", err)
		}

		msg = ItemExchangeMessage{
			Message: Message{
				MessageType:   msgType,
				TransactionID: tranID,
			},
			OfferedItem: int64(data["offeredItem"].(float64)),
			WantedItem:  int64(data["wantedItem"].(float64)),
			SenderID:    data["senderId"].(int64),
		}
	case NodesList:
		nodes := make(map[int64]string)
		if nodesData, ok := data["nodes"].(map[string]interface{}); ok {
			for idStr, addr := range nodesData {
				id, err := strconv.ParseInt(idStr, 10, 64)
				if err != nil {
					log.Printf("Failed to parse node ID: %s, error: %v", idStr, err)
					continue
				}
				if addrStr, ok := addr.(string); ok {
					nodes[id] = addrStr
				} else {
					log.Printf("Invalid address format for node ID: %d", id)
				}
			}
		}

		var nodeList []string
		for _, addr := range nodes {
			nodeList = append(nodeList, addr)
		}

		msg = NodesListMessage{
			Message: Message{
				MessageType:   msgType,
				TransactionID: tranID,
			},
			SenderID:  int64(data["senderId"].(float64)),
			Nodes:    nodeList,
		}
	case NodesListResponse:
		rawNeighbours, ok := data["neighbour"].(map[string]interface{})
		if !ok {
			log.Println("Invalid neighbour map in data:", data)
			return nil, fmt.Errorf("invalid neighbour map")
		}

		neighbours := make(map[int64]neighbour.Neighbour)

		for idStr, raw := range rawNeighbours {
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				log.Println("Failed to parse neighbour ID:", err)
				continue
			}

			rawMap, ok := raw.(map[string]interface{})
			if !ok {
				log.Println("Invalid neighbour entry:", raw)
				continue
			}

			n, err := neighbour.NeighbourFromMap(rawMap)
			if err != nil {
				log.Println("Failed to parse neighbour:", err)
				continue
			}

			neighbours[id] = n
		}

		msg = NodesListResponseMessage{
			Message: Message{
				MessageType:   msgType,
				TransactionID: tranID,
			},
			SenderID:  int64(data["senderId"].(float64)),
			ParticipatingNodes: neighbours,
		}
	case Confirm:
		msg = ConfirmMessage{
			Message: Message{
				MessageType:   msgType,
				TransactionID: tranID,
			},
			SenderID:  int64(data["senderId"].(float64)),
		}
	case GetItems:
		msg = GetItemsMessage{
			Message: Message{
				MessageType:   msgType,
				TransactionID: tranID,
			},
			SenderID:  int64(data["senderId"].(float64)),
		}

	case ErrorType:
		msg = ErrorMessage{
			Message: Message{
				MessageType:   msgType,
				TransactionID: tranID,
			},
		}
	case AnnounceSelf:
		var listeningAddr serializers.MultiaddrJSON
		if err := listeningAddr.UnmarshalJSON([]byte(data["listeningAddress"].(string))); err != nil {
			log.Println("Failed to unmarshal listening address:", err)
			return nil, fmt.Errorf("failed to unmarshal listening address: %w", err)
		}

		msg = AnnounceSelfMessage{
			Message: Message{
				MessageType:   msgType,
				TransactionID: tranID,
			},
			ID:               int64(data["id"].(float64)),
			ListeningAddress: listeningAddr,
		}
	default:
		log.Println("Unknown message type:", msgType)
		return nil, fmt.Errorf("unknown message type: %v", msgType)
	}

	return msg, nil
}
