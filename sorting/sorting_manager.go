package sorting

import (
	"fmt"
	"log"
	"sort"
	"sync"

	"github.com/cpprian/distributed-sort-golang/messages"
	"github.com/cpprian/distributed-sort-golang/serializers"
	"github.com/libp2p/go-libp2p/core/peer"

	ma "github.com/multiformats/go-multiaddr"
)

// HostInterface defines the methods required for the host in SortingManager.
type HostInterface interface {
	GetListenAddress() string
	Broadcast(interface{})
	SendMessage(msg messages.MessageInterface)
	GetPeerID() peer.ID
	GetAddrs() []ma.Multiaddr
}

type SortingManager struct {
	ID                 int64
	Items              []int64
	ParticipatingNodes map[int64]string // Maps node ID to address
	Host               HostInterface
	mu                 sync.Mutex
}

func NewSortingManager() *SortingManager {
	id := int64(0)

	log.Printf("Creating SortingManager with ID: %d", id)

	return &SortingManager{
		ID:                 id,
		Items:              []int64{},
		ParticipatingNodes: make(map[int64]string),
		Host:               nil,
		mu:                 sync.Mutex{},
	}
}

func (sm *SortingManager) SetHost(host HostInterface) {
	sm.mu.Lock()
	sm.Host = host
	sm.mu.Unlock()
}

func (sm *SortingManager) AnnounceSelf() {
	log.Println("Announcing self with ID:", sm.ID)

	addrInfo := peer.AddrInfo{
		ID:    sm.Host.GetPeerID(),
		Addrs: sm.Host.GetAddrs(),
	}

	addrs, err := peer.AddrInfoToP2pAddrs(&addrInfo)
	if err != nil || len(addrs) == 0 {
		log.Println("Could not get proper full multiaddr:", err)
		return
	}

	addrStr := addrs[0].String()
	log.Printf("Announcing self with ID %d at address %s", sm.ID, addrStr)

	addr, err := ma.NewMultiaddr(addrStr)
	if err != nil {
		log.Printf("Failed to parse listen address: %v", err)
		return
	}

	msg := messages.NewAnnounceSelfMessage(sm.ID, serializers.MultiaddrJSON{Multiaddr: addr})
	sm.Host.Broadcast(msg)
}

func (sm *SortingManager) Add(item int64) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	log.Println("Adding item:", item)
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

	log.Printf("Adding item: %d. Local items before: %v", item, sm.Items)
	sm.Items = append(sm.Items, item)
	intItems := make([]int, len(sm.Items))
	for i, v := range sm.Items {
		intItems[i] = int(v)
	}
	sort.Ints(intItems)
	for i, v := range intItems {
		sm.Items[i] = int64(v)
	}
}

func (sm *SortingManager) Remove(item int64) {
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

func (sm *SortingManager) SendCornerChange(direction string, item int64) {
	log.Printf("Sending corner item change: item %d, direction %s, from %d", item, direction, sm.ID)
	msg := messages.NewCornerItemChangeMessage(item, int64(sm.ID), direction)
	sm.Host.Broadcast(msg)
}

func (sm *SortingManager) OrderItemsExchange(offeredItem, wantedItem int64, neighbourID int64, transactionID string) {
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

	log.Printf("Responding to ItemExchange: offered %d, wanted %d, from %d", offeredItem, wantedItem, senderID)

	var itemToSend int64
	if senderID > sm.ID {
		log.Println("Sender ID is greater than local ID, sending from the right end")
		if len(sm.Items) == 0 || sm.Items[len(sm.Items)-1] != wantedItem {
			sm.Host.SendMessage(messages.NewErrorMessage(transactionID))
			return
		}
		itemToSend = sm.Items[len(sm.Items)-1]
		sm.Items[len(sm.Items)-1] = offeredItem
	} else {
		log.Println("Sender ID is less than or equal to local ID, sending from the left end")
		if len(sm.Items) == 0 || sm.Items[0] != wantedItem {
			sm.Host.SendMessage(messages.NewErrorMessage(transactionID))
			return
		}
		itemToSend = sm.Items[0]
		sm.Items[0] = offeredItem
	}
	intItems := make([]int, len(sm.Items))
	for i, v := range sm.Items {
		intItems[i] = int(v)
	}
	sort.Ints(intItems)
	for i, v := range intItems {
		sm.Items[i] = int64(v)
	}

	response := messages.ItemExchangeMessage{
		Message: messages.Message{
			MessageType:   messages.ItemExchange,
			TransactionID: transactionID,
		},
		OfferedItem: itemToSend,
		WantedItem:  wantedItem,
		SenderID:    int64(sm.ID),
	}

	log.Printf("Sending ItemExchange response: offered %d, wanted %d, to %d", itemToSend, wantedItem, senderID)
	sm.Host.SendMessage(response)
}

func (sm *SortingManager) ProcessCornerItemChange(msg messages.CornerItemChangeMessage) {
	fmt.Printf("Processing CornerItemChange: item %d, direction %s, from %d\n",
		msg.Item, msg.Direction, msg.SenderID)
	confirm := messages.ConfirmMessage{
		TransactionID: msg.TransactionID,
		SenderID:      int64(sm.ID),
	}
	sm.Host.SendMessage(confirm)
}

func (sm *SortingManager) Activate(knownParticipant string) {
	if knownParticipant == "" {
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

		sm.ParticipatingNodes[sm.ID] = m.String()
		log.Println("All known participants:", len(sm.ParticipatingNodes))
		sm.ID = maxKey(sm.ParticipatingNodes) + 1
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
		sm.Host.Broadcast(msg)
	}
}

func maxKey(m map[int64]string) int64 {
	var max int64 = -1
	for k := range m {
		if k > max {
			max = k
		}
	}
	return max
}

func (sm *SortingManager) GetItems() []int64 {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	return sm.Items
}

func (sm *SortingManager) GetAllItems() []int64 {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	itemsCopy := make([]int64, len(sm.Items))
	copy(itemsCopy, sm.Items)
	return itemsCopy
}

func (sm *SortingManager) GetParticipatingNodes() map[int64]string {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	nodesCopy := make(map[int64]string)
	for k, v := range sm.ParticipatingNodes {
		nodesCopy[int64(k)] = v
	}
	return nodesCopy
}
