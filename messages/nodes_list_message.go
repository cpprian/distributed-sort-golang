package messages

// NodesListMessage represents a message containing a list of nodes.
type NodesListMessage struct {
	Nodes []string `json:"nodes"`
}