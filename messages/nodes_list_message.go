package messages

type NodesListMessage struct {
	BaseMessage
}

func NewNodesListMessage() NodesListMessage {
	return NodesListMessage{
		NewBaseMessage(NodesList),
	}
}
