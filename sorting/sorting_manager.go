package sorting

import (
	"fmt"
	"log"
	"sort"
	"sync"

	"github.com/cpprian/distributed-sort-golang/messages"
	"github.com/cpprian/distributed-sort-golang/serializers"

	ma "github.com/multiformats/go-multiaddr"
)

// HostInterface defines the methods required for the host in SortingManager.
type HostInterface interface {
	GetListenAddress() string
	Broadcast(msg messages.MessageInterface, msgType messages.MessageType)
	SendMessage(msg messages.MessageInterface)
}
type SortingManager struct {
	ID                 int
	Items              []int
	ParticipatingNodes map[int]string // Maps node ID to address
	Host               HostInterface
	mu                 sync.Mutex
}

func NewSortingManager() *SortingManager {
	return &SortingManager{
		ID:                 0,
		Items:              []int{},
		ParticipatingNodes: make(map[int]string),
	}
}

func (sm *SortingManager) SetHost(host HostInterface) {
	sm.Host = host
	go sm.AnnounceSelf()
}

func (sm *SortingManager) AnnounceSelf() {
	addrStr := sm.Host.GetListenAddress()

	addr, err := ma.NewMultiaddr(addrStr)
	if err != nil {
		log.Printf("Failed to parse listen address: %v", err)
		return
	}

	msg := messages.NewAnnounceSelfMessage(int64(sm.ID), serializers.MultiaddrJSON{Multiaddr: addr})
	log.Printf("Announcing self with ID %d at address %s", sm.ID, addrStr)

	// Broadcast the message to all nodes
	sm.Host.Broadcast(msg.Message, msg.Type())
}

func (sm *SortingManager) Add(item int) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if len(sm.Items) > 0 {
		if item < sm.Items[0] {
			sm.SendCornerChange("left", item)
		}
		if item > sm.Items[len(sm.Items)-1] {
			sm.SendCornerChange("right", item)
		}
	} else {
		sm.SendCornerChange("left", item)
		sm.SendCornerChange("right", item)
	}
	sm.Items = append(sm.Items, item)
	sort.Ints(sm.Items)
}

func (sm *SortingManager) Remove(item int) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for i, v := range sm.Items {
		if v == item {
			sm.Items = append(sm.Items[:i], sm.Items[i+1:]...)
			log.Printf("Item removed. Local items: %v", sm.Items)
			return
		}
	}
}

func (sm *SortingManager) SendCornerChange(direction string, item int) {
	msg := messages.CornerItemChangeMessage{
		Item:     item,
		SenderID: int64(sm.ID),
	}
	sm.Host.Broadcast(msg, msg.Type())
}

func (sm *SortingManager) OrderItemsExchange(offeredItem, wantedItem int, neighbourID int, transactionID string) {
	msg := messages.ItemExchangeMessage{
		OfferedItem: offeredItem,
		WantedItem:  wantedItem,
		SenderID:    int64(sm.ID),
	}
	sm.Remove(offeredItem)
	sm.Host.SendMessage(msg)
}

func (sm *SortingManager) RespondToItemsExchange(msg messages.ItemExchangeMessage) {
	transactionID := msg.TransactionID
	senderID := msg.SenderID
	wantedItem := msg.WantedItem
	offeredItem := msg.OfferedItem

	sm.mu.Lock()
	defer sm.mu.Unlock()

	var itemToSend int
	if senderID > int64(sm.ID) {
		if len(sm.Items) == 0 || sm.Items[len(sm.Items)-1] != wantedItem {
			sm.Host.SendMessage(messages.NewErrorMessage(transactionID))
			return
		}
		itemToSend = sm.Items[len(sm.Items)-1]
		sm.Items[len(sm.Items)-1] = offeredItem
	} else {
		if len(sm.Items) == 0 || sm.Items[0] != wantedItem {
			sm.Host.SendMessage(messages.NewErrorMessage(transactionID))
			return
		}
		itemToSend = sm.Items[0]
		sm.Items[0] = offeredItem
	}
	sort.Ints(sm.Items)

	response := messages.ItemExchangeMessage{
		OfferedItem: itemToSend,
		WantedItem:  wantedItem,
		SenderID:    int64(sm.ID),
	}
	sm.Host.SendMessage(response.Message)
}

func (sm *SortingManager) ProcessMessage(msg messages.MessageInterface) {
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
				sm.Items = append([]int{m.OfferedItem}, sm.Items...)
			}
			sort.Ints(sm.Items)
			sm.mu.Unlock()
		} else {
			sm.RespondToItemsExchange(m)
		}
	case messages.AnnounceSelf:
		m := msg.(messages.AnnounceSelfMessage)
		sm.ParticipatingNodes[int(m.ID)] = m.ListeningAddress.String()
		sm.mu.Lock()
		defer sm.mu.Unlock()
		if sm.ID < int(m.ID) {
			sm.ID = int(m.ID)
		}
		sm.Host.Broadcast(messages.NewAnnounceSelfMessage(int64(sm.ID), serializers.MultiaddrJSON{Multiaddr: ma.StringCast(sm.Host.GetListenAddress())}).Message, messages.AnnounceSelf)
		// Log the current participating nodes
		log.Printf("Participating nodes updated: %v", sm.ParticipatingNodes)
		// Notify that self has been announced
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
			if _, exists := sm.ParticipatingNodes[id]; !exists {
				sm.ParticipatingNodes[id] = addr
			}
		}
		log.Printf("Updated participating nodes: %v", sm.ParticipatingNodes)
	case messages.NodesListResponse:
		m := msg.(messages.NodesListResponseMessage)
		fmt.Printf("Received NodesListResponse from sender ID: %d with nodes: %v\n", m.SenderID, m.ParticipatingNodes)
		sm.mu.Lock()
		defer sm.mu.Unlock()
		for id, addr := range m.ParticipatingNodes {
			if _, exists := sm.ParticipatingNodes[int(id)]; !exists {
				sm.ParticipatingNodes[int(id)] = addr.String()	
			}
		}
		log.Printf("Updated participating nodes from response: %v", sm.ParticipatingNodes)
	default:
		fmt.Println("Received unknown message:", msg)
	}
}

func (sm *SortingManager) ProcessCornerItemChange(msg messages.CornerItemChangeMessage) {
	fmt.Printf("Processing CornerItemChange: item %d, direction %s, from %d\n",
		msg.Item, msg.Direction, msg.SenderID)
	confirm := messages.ConfirmMessage{TransactionID: msg.TransactionID}
	sm.Host.SendMessage(confirm)
}

func (sm *SortingManager) Activate(knownParticipant string) {
	if knownParticipant == "" {
		sm.ID = 0
		sm.ParticipatingNodes[sm.ID] = sm.Host.GetListenAddress()
		log.Printf("Activated SortingManager with ID %d and no known participants", sm.ID)
		if sm.Host == nil {
			log.Println("Host is not set. Cannot announce self.")
			return
		}
		sm.AnnounceSelf()
	} else {
		// If a known participant is provided, set the ID to the maximum existing ID + 1
		// and add the known participant to the ParticipatingNodes map.
		// This is a simplified example; in a real scenario, you would fetch the existing nodes
		// and determine the ID based on the maximum existing node ID.
		sm.mu.Lock()
		defer sm.mu.Unlock()
		m, err := ma.NewMultiaddr(knownParticipant)
		if err != nil {
			log.Printf("Failed to parse known participant address: %v", err)
			return
		}

		sm.ID = maxKey(sm.ParticipatingNodes) + 1
		sm.ParticipatingNodes[sm.ID] = m.String()
		log.Printf("Activated SortingManager with ID %d and known participant %s", sm.ID, knownParticipant)

		if sm.Host == nil {
			log.Println("Host is not set. Cannot announce self.")
			return
		}

		// Announce self with the new ID and address
		addrStr := sm.Host.GetListenAddress()
		addr, err := ma.NewMultiaddr(addrStr)
		if err != nil {
			log.Printf("Failed to parse listen address: %v", err)
			return
		}
		msg := messages.NewAnnounceSelfMessage(int64(sm.ID), serializers.MultiaddrJSON{Multiaddr: addr})
		log.Printf("Announcing self with ID %d at address %s", sm.ID, addrStr)
		sm.Host.Broadcast(msg.Message, msg.Type())
	}
}

func maxKey(m map[int]string) int {
	max := -1
	for k := range m {
		if k > max {
			max = k
		}
	}
	return max
}

func (sm *SortingManager) GetItems() []int {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	return sm.Items
}

func (sm *SortingManager) GetAllItems() []int {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	itemsCopy := make([]int, len(sm.Items))
	copy(itemsCopy, sm.Items)
	return itemsCopy
}

func (sm *SortingManager) GetParticipatingNodes() map[int]string {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	nodesCopy := make(map[int]string)
	for k, v := range sm.ParticipatingNodes {
		nodesCopy[k] = v
	}
	return nodesCopy
}


