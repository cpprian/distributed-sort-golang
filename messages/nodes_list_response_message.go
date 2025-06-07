package messages

import (
	"github.com/cpprian/distributed-sort-golang/neighbours"
	"github.com/google/uuid"
)

type NodesListResponseMessage struct {
	BaseMessage
	ParticipatingNodes map[int64]neighbours.Neighbour `json:"participatingNodes"` // Maps node ID to Neighbour
}

func NewNodesListResponseMessage(participatingNodes map[int64]neighbours.Neighbour, transactionID uuid.UUID) NodesListResponseMessage {
	return NodesListResponseMessage{
		BaseMessage:        NewBaseMessageWithTransactionID(NodesListResponse, transactionID),
		ParticipatingNodes: participatingNodes,
	}
}
