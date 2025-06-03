package distributed_sort

import (
	"fmt"
	"log"
	"sort"
	"sync"

	"github.com/cpprian/distributed-sort-golang/messages"
	"github.com/cpprian/distributed-sort-golang/serializers"
)

// HostInterface defines the methods required for the host in SortingManager.
type HostInterface interface {
	GetListenAddress() string
	Broadcast(msg messages.Message, msgType string)
	SendMessage(msg messages.Message)
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
	msg := messages.AnnounceSelfMessage{
		ID:               int64(sm.ID),
		ListeningAddress: serializers.MultiaddrJSON(sm.Host.GetListenAddress()),
	}
	sm.Host.Broadcast(msg, msg.Type())
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
	msg := messages.CornerItemChangeMessage[int]{
		Item:      item,
		SenderID:  int64(sm.ID),
	}
	sm.Host.Broadcast(msg, msg.Type())
}

func (sm *SortingManager) OrderItemsExchange(offeredItem, wantedItem int, neighbourID int, transactionID string) {
	msg := messages.ItemExchangeMessage[int]{
		OfferedItem:   offeredItem,
		WantedItem:    wantedItem,
		SenderID:      int64(sm.ID),
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
	if senderID > sm.ID {
		if len(sm.Items) == 0 || sm.Items[len(sm.Items)-1] != wantedItem {
			sm.Host.SendMessage(ErrorMessage{
				Error:         "Wanted item mismatch",
				TransactionID: transactionID,
			})
			return
		}
		itemToSend = sm.Items[len(sm.Items)-1]
		sm.Items[len(sm.Items)-1] = offeredItem
	} else {
		if len(sm.Items) == 0 || sm.Items[0] != wantedItem {
			sm.Host.SendMessage(ErrorMessage{
				Error:         "Wanted item mismatch",
				TransactionID: transactionID,
			})
			return
		}
		itemToSend = sm.Items[0]
		sm.Items[0] = offeredItem
	}
	sort.Ints(sm.Items)

	response := messages.ItemExchangeMessage{
		OfferedItem:   itemToSend,
		WantedItem:    wantedItem,
		SenderID:      sm.ID,
		TransactionID: transactionID,
	}
	sm.Host.SendMessage(response)
}

func (sm *SortingManager) ProcessMessage(msg Message) {
	switch m := msg.(type) {
	case CornerItemChangeMessage:
		sm.ProcessCornerItemChange(m)
	case ItemExchangeMessage:
		if m.Response {
			fmt.Println("Received ItemExchange response:", m)
			sm.mu.Lock()
			if m.SenderID > sm.ID {
				sm.Items = append(sm.Items, m.OfferedItem)
			} else {
				sm.Items = append([]int{m.OfferedItem}, sm.Items...)
			}
			sort.Ints(sm.Items)
			sm.mu.Unlock()
		} else {
			sm.RespondToItemsExchange(m)
		}
	case AnnounceSelfMessage:
		sm.ParticipatingNodes[m.ID] = m.ListeningAddress
		fmt.Println("AnnounceSelf received. Nodes now:", sm.ParticipatingNodes)
	case GetItemsMessage:
		sm.Host.SendMessage(GetItemsMessage{Items: sm.Items})
	default:
		fmt.Println("Received unknown message:", msg)
	}
}

func (sm *SortingManager) ProcessCornerItemChange(msg CornerItemChangeMessage) {
	fmt.Printf("Processing CornerItemChange: item %d, direction %s, from %d\n",
		msg.Item, msg.Direction, msg.SenderID)
	confirm := ConfirmMessage{TransactionID: msg.TransactionID}
	sm.Host.SendMessage(confirm)
}

func (sm *SortingManager) Activate(knownParticipant string) {
	if knownParticipant == "" {
		sm.ID = 0
		sm.ParticipatingNodes[sm.ID] = sm.Host.GetListenAddress()
	} else {
		// Simulate waiting for population
		// In real code, perform a request to knownParticipant to get node list
		if len(sm.ParticipatingNodes) > 0 {
			sm.ID = maxKey(sm.ParticipatingNodes) + 1
		} else {
			sm.ID = 0
		}
		sm.ParticipatingNodes[sm.ID] = sm.Host.GetListenAddress()
	}
	sm.AnnounceSelf()
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
