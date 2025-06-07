package messages

// NodesListMessage represents a message containing a list of nodes.
type NodesListMessage struct {
	Message
	Nodes    []string `json:"nodes"`
	SenderID int64    `json:"sender_id"`
}
