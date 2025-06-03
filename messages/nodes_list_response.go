package messages

import (
	"github.com/cpprian/distributed-sort-golang/neighbour"
	"github.com/google/uuid"
)

type NodesListResponseMessage struct {
	Message
	ParticipatingNodes map[int64]neighbour.Neighbour `json:"participatingNodes"`
	SenderID           int64                         `json:"sender_id"`
}

func NewNodesListResponseMessage(nodes map[int64]neighbour.Neighbour) NodesListResponseMessage {
	return NodesListResponseMessage{
		Message:            NewMessage(NodesListResponse),
		ParticipatingNodes: nodes,
	}
}

func NewNodesListResponseMessageWithID(nodes map[int64]neighbour.Neighbour, txID uuid.UUID) NodesListResponseMessage {
	return NodesListResponseMessage{
		Message:            Message{MessageType: NodesListResponse, TransactionID: txID},
		ParticipatingNodes: nodes,
	}
}

func (m NodesListResponseMessage) Type() MessageType {
	return NodesListResponse
}
