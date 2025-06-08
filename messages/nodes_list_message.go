package messages

import "github.com/google/uuid"

type NodesListMessage struct {
	BaseMessage
}

func NewNodesListMessage() NodesListMessage {
	return NodesListMessage{
		NewBaseMessage(NodesList),
	}
}


func (m NodesListMessage) GetMessageType() MessageType {
	return NodesList
}

func (m NodesListMessage) GetTransactionID() uuid.UUID {
	return m.TransactionID
}
