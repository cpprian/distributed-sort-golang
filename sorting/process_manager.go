package sorting

import (
	"fmt"
	"log"

	"github.com/cpprian/distributed-sort-golang/messages"
)

func (sm *SortingManager) ProcessMessage(msg messages.MessageInterface) {
	switch msg.Type() {
	case messages.CornerItemChange:
		m := msg.(messages.CornerItemChangeMessage)
		sm.ProcessCornerItemChange(m)
	case messages.AnnounceSelf:
		m := msg.(messages.AnnounceSelfMessage)
		sm.mu.Lock()
		defer sm.mu.Unlock()
		sm.ParticipatingNodes[m.ID] = m.ListeningAddress.String()
		fmt.Printf("Node %d announced itself with address %s. Current participating nodes: %v\n", m.ID, m.ListeningAddress, sm.ParticipatingNodes)
	case messages.GetItems:
		m := msg.(messages.GetItemsMessage)
		sm.Host.SendMessage(messages.GetItemsMessage{Items: sm.Items})
		fmt.Printf("GetItems request received from sender ID: %d. Current items: %v\n", m.SenderID, sm.Items)
	case messages.NodesList:
		m := msg.(messages.NodesListMessage)
		fmt.Printf("Received NodesList from sender ID: %d with nodes: %v\n", m.SenderID, m.Nodes)
		sm.mu.Lock()
		defer sm.mu.Unlock()
		for id, addr := range m.Nodes {
			if _, exists := sm.ParticipatingNodes[int64(id)]; !exists {
				sm.ParticipatingNodes[int64(id)] = addr
			}
		}
		log.Printf("Updated participating nodes: %v", sm.ParticipatingNodes)
	default:
		fmt.Println("Received unknown message:", msg)
	}
}
