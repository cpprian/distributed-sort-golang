package sorting

import (
	"log"
	"sort"
	"sync"
	"time"

	"github.com/cpprian/distributed-sort-golang/messages"
	"github.com/cpprian/distributed-sort-golang/neighbours"
	"github.com/cpprian/distributed-sort-golang/networking"
	"github.com/cpprian/distributed-sort-golang/utils"
	ma "github.com/multiformats/go-multiaddr"
)

type SortingManager struct {
	ID                 int64
	Items              []int64
	ParticipatingNodes map[int64]neighbours.Neighbour
	Host               networking.Libp2pHost
	Self               *neighbours.Neighbour
	Messaging          networking.MessagingController
	mu                 sync.Mutex
}

func NewSortingManager(host networking.Libp2pHost, messagingController networking.MessagingController) *SortingManager {
	return &SortingManager{
		ID:                 0,
		Items:              []int64{},
		ParticipatingNodes: make(map[int64]neighbours.Neighbour),
		Host:               host,
		Self:               neighbours.NewNeighbour(host.Host.Network().ListenAddresses()[0], 0),
		Messaging:          messagingController,
		mu:                 sync.Mutex{},
	}
}

func (sm *SortingManager) Activate(knownParticipant ma.Multiaddr) {
	log.Println("Activating SortingManager...")
	sm.ParticipatingNodes[sm.Self.ID] = *sm.Self
	sm.Self.ID = sm.ID

	if knownParticipant == nil {
		log.Printf("No known participant. Setting self ID to %d and address to %s\n", sm.Self.ID, sm.Self.Multiaddr.String())
	} else {
		log.Println("Retrieving participating nodes from known participant:", knownParticipant)
		nodes, err := sm.Messaging.RetrieveParticipatingNodes(sm.Host.Host, knownParticipant, sm.Messaging.GetProtocolID(), sm.Messaging.GetMessageProcessor())
		if err != nil {
			log.Println("Error retrieving participating nodes: ", err)
			return
		}

		sm.ParticipatingNodes = nodes
		sm.ID = utils.MaxKey(nodes) + 1
		log.Printf("Setting self ID to %d and address to %s\n", sm.Self.ID, sm.Self.Multiaddr.String())

		sm.AnnounceSelf()
	}
}

func (sm *SortingManager) AnnounceSelf() {
	log.Println("Announcing self with ID: ", sm.ID)
	
	for id, neighbour := range sm.ParticipatingNodes {
		if id == sm.ID {
			continue
		}

		log.Println("Announcing self to neighbour: ", neighbour.Multiaddr.String())

		controller, err := networking.DialByMultiaddr(sm.Host.Host, neighbour.Multiaddr, sm.Messaging.GetProtocolID(), sm.Messaging.GetMessageProcessor())
		if err != nil {
			log.Printf("Failed to dial neighbour %s: %v", neighbour.Multiaddr.String(), err)
			continue
		}

		msg := messages.NewAnnounceSelfMessage(sm.ID, sm.Self.Multiaddr)
		go func(c networking.MessagingController, m messages.IMessage) {
			defer c.Close()
			
			select {
			case <-c.SendMessage(m):
				log.Printf("Sent AnnounceSelf to %v\n", neighbour)
			case <-time.After(3 * time.Second):
				log.Printf("Timeout while sending AnnounceSelf to %v\n", neighbour)
			}
		}(controller, msg)
	}

	log.Println("Self announced successfully. Current participating nodes:", len(sm.ParticipatingNodes))
}

func (sm *SortingManager) RemoveItem(item int64) {
	log.Println("Removing item: ", item)
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for i, v := range sm.Items {
		if v == item {
			sm.Items = append(sm.Items[:i], sm.Items[i+1:]...)
			log.Println("Item removed successfully.")
			return
		}
	}
	log.Println("Item not found in the list.")
}

func (sm *SortingManager) GetFirstItem() int64 {
	return sm.Items[0]
}

func (sm *SortingManager) GetLastItem() int64 {
	if len(sm.Items) == 0 {
		return 0
	}
	return sm.Items[len(sm.Items)-1]
}

func (sm *SortingManager) getNeighbour(isRight bool) *neighbours.Neighbour {
	var (
		targetID  int64
		found     bool
		neighbour neighbours.Neighbour
	)

	for id, n := range sm.ParticipatingNodes {
		if (isRight && id > sm.ID) || (!isRight && id < sm.ID) {
			if !found || (isRight && id < targetID) || (!isRight && id > targetID) {
				targetID = id
				neighbour = n
				found = true
			}
		}
	}

	if found {
		return &neighbour
	}
	return nil
}

func (sm *SortingManager) GetRightNeighbour() *neighbours.Neighbour {
	return sm.getNeighbour(true)
}

func (sm *SortingManager) GetLeftNeighbour() *neighbours.Neighbour {
	return sm.getNeighbour(false)
}

func (sm *SortingManager) GetAllItems() []int64 {
	log.Println("Retrieving all items from participating nodes...")
	var allItems []int64

	ids := make([]int64, 0, len(sm.ParticipatingNodes))
	for id := range sm.ParticipatingNodes {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool {
		return ids[i] < ids[j]
	})

	for _, id := range ids {
		neighbour := sm.ParticipatingNodes[id]
		log.Printf("Retrieving items from neighbour %d at %s\n", id, neighbour.Multiaddr.String())

		controller, err := networking.DialByMultiaddr(sm.Host.Host, neighbour.Multiaddr, sm.Messaging.GetProtocolID(), sm.Messaging.GetMessageProcessor())
		if err != nil {
			log.Printf("Failed to dial neighbour %s: %v", neighbour.Multiaddr.String(), err)
			continue
		}

		msg := messages.NewGetItemsMessage()
		responseChan := controller.SendMessage(msg)
		select {
		case response := <-responseChan:
			if itemsMsg, ok := response.(messages.GetItemsMessage); ok {
				log.Printf("Received items from neighbour %d: %v", id, itemsMsg.Items)
				allItems = append(allItems, itemsMsg.Items...)
			} else {
				log.Printf("Unexpected response type from neighbour %d: %T", id, response)
			}
		case <-time.After(3 * time.Second):
			log.Printf("Timeout while waiting for items from neighbour %d", id)
		}

		controller.Close()
		log.Printf("Finished retrieving items from neighbour %d\n", id)
	}

	return allItems
}

func (sm *SortingManager) ProcessMessage(msg messages.IMessage, controller networking.MessagingController) {
	log.Printf("Processing message of type %T from %s\n", msg, controller.GetRemoteAddress())

	switch m := msg.(type) {
	case messages.CornerItemChangeMessage: 
		log.Printf("Received CornerItemChangeMessage: %s", msg)
		sm.ProcessCornerItemChange(m, controller)
	case messages.NodesListMessage:
		log.Println("Received NodesListMessage, updating participating nodes...")
		response := messages.NewNodesListResponseMessage(sm.ParticipatingNodes, m.GetTransactionID())
		controller.SendMessage(response)
	case messages.AnnounceSelfMessage:
		log.Printf("Received AnnounceSelfMessage from %d: %s", m.ID, m.ListeningAddress)
		neighbour := neighbours.NewNeighbour(m.ListeningAddress, m.ID)
		sm.mu.Lock()
		defer sm.mu.Unlock()
		sm.ParticipatingNodes[m.ID] = *neighbour
	case messages.GetItemsMessage:
		log.Println("Received GetItemsMessage, sending items...")
		sm.mu.Lock()
		defer sm.mu.Unlock()
		response := messages.NewGetItemsMessageWithTransactionID(sm.Items, m.GetTransactionID())
		controller.SendMessage(response)
	default:
		log.Printf("Unknown message type: %T\n", msg)
	}

	log.Println("Message processed successfully.")
}