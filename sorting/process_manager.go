package sorting

import (
	"fmt"
	"log"
	"sort"

	"github.com/cpprian/distributed-sort-golang/messages"
	"github.com/cpprian/distributed-sort-golang/serializers"

	ma "github.com/multiformats/go-multiaddr"
)

func (sm *SortingManager) ProcessMessage(msg messages.MessageInterface) {
	log.Println("Processing message:", msg)

	switch msg.Type() {
	case messages.CornerItemChange:
		m := msg.(messages.CornerItemChangeMessage)
		sm.ProcessCornerItemChange(m)
	case messages.ItemExchange:
		m := msg.(messages.ItemExchangeMessage)
		if m.Response {
			fmt.Println("Received ItemExchange response:", m)
			sm.mu.Lock()
			if m.SenderID > int64(sm.ID) {
				sm.Items = append(sm.Items, m.OfferedItem)
			} else {
				sm.Items = append([]int64{m.OfferedItem}, sm.Items...)
			}
			intItems := make([]int, len(sm.Items))
			for i, v := range sm.Items {
				intItems[i] = int(v)
			}
			sort.Ints(intItems)
			for i, v := range intItems {
				sm.Items[i] = int64(v)
			}
			sm.mu.Unlock()
		} else {
			sm.RespondToItemsExchange(m)
		}
	case messages.AnnounceSelf:
		m := msg.(messages.AnnounceSelfMessage)
		sm.ParticipatingNodes[m.ID] = m.ListeningAddress.String()
		sm.mu.Lock()
		defer sm.mu.Unlock()
		if sm.ID < m.ID {
			sm.ID = m.ID
		}
		sm.Host.Broadcast(messages.NewAnnounceSelfMessage(int64(sm.ID), serializers.MultiaddrJSON{Multiaddr: ma.StringCast(sm.Host.GetListenAddress())}))
		fmt.Println("AnnounceSelf received. Nodes now:", sm.ParticipatingNodes)
	case messages.GetItems:
		m := msg.(messages.GetItemsMessage)
		sm.Host.SendMessage(messages.GetItemsMessage{Items: sm.Items})
		fmt.Printf("GetItems request received from sender ID: %d. Current items: %v\n", m.SenderID, sm.Items)
	case messages.ErrorType:
		m := msg.(messages.ErrorMessage)
		fmt.Println("Received error message:", m)
	case messages.Confirm:
		m := msg.(messages.ConfirmMessage)
		fmt.Printf("Received confirmation for transaction ID: %s from sender ID: %d\n", m.TransactionID, m.SenderID)
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
	case messages.NodesListResponse:
		m := msg.(messages.NodesListResponseMessage)
		fmt.Printf("Received NodesListResponse from sender ID: %d with nodes: %v\n", m.SenderID, m.ParticipatingNodes)
		sm.mu.Lock()
		defer sm.mu.Unlock()
		for id, addr := range m.ParticipatingNodes {
			if _, exists := sm.ParticipatingNodes[id]; !exists {
				sm.ParticipatingNodes[id] = addr.String()
			}
		}
		log.Printf("Updated participating nodes from response: %v", sm.ParticipatingNodes)
	default:
		fmt.Println("Received unknown message:", msg)
	}
}
