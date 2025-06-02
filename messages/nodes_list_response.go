package messages

import (
	"github.com/google/uuid"
	"github.com/cpprian/distributed-sort-golang"
)

type NodesListResponseMessage struct {
	Message
	ParticipatingNodes map[int64]distributed_sort.Neighbour `json:"participatingNodes"`
}

func NewNodesListResponseMessage(nodes map[int64]distributed_sort.Neighbour) NodesListResponseMessage {
	return NodesListResponseMessage{
		Message:            NewMessage(NodesListResponse),
		ParticipatingNodes: nodes,
	}
}

func NewNodesListResponseMessageWithID(nodes map[int64]distributed_sort.Neighbour, txID uuid.UUID) NodesListResponseMessage {
	return NodesListResponseMessage{
		Message:            Message{MessageType: NodesListResponse, TransactionID: txID},
		ParticipatingNodes: nodes,
	}
}
