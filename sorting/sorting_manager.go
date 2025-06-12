package sorting

import (
	"log"
	"sort"
	"sync"
	"time"

	"github.com/cpprian/distributed-sort-golang/messages"
	"github.com/cpprian/distributed-sort-golang/neighbours"
	"github.com/cpprian/distributed-sort-golang/networks"
	"github.com/cpprian/distributed-sort-golang/utils"
	peer "github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
)

type SortingManager struct {
	ID                 int64
	Items              []int64
	ParticipatingNodes map[int64]neighbours.Neighbour
	Host               *networks.Libp2pHost
	Self               *neighbours.Neighbour
	Messaging          networks.MessagingController
	mu                 sync.Mutex
}

func NewSortingManager() *SortingManager {
	return &SortingManager{
		ID:                 0,
		Items:              []int64{},
		ParticipatingNodes: make(map[int64]neighbours.Neighbour),
		Host:               networks.NewEmptyLibp2pHost(),
		Self:               nil,
		Messaging:          nil,
		mu:                 sync.Mutex{},
	}
}

func (sm *SortingManager) SetHost(host *networks.Libp2pHost) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.Host = host

	addrInfo := peer.AddrInfo{
		ID:    host.Host.ID(),
		Addrs: host.Host.Addrs(),
	}
	fullAddrs, err := peer.AddrInfoToP2pAddrs(&addrInfo)
	if err != nil || len(fullAddrs) == 0 {
		log.Fatalf("Failed to convert to full multiaddr: %v", err)
	}
	sm.Self = neighbours.NewNeighbour(fullAddrs[0], 0)
	log.Printf("Host set with address: %s and ID: %d\n", sm.Host.Host.Addrs()[0], sm.Self.ID)
}

func (sm *SortingManager) SetMessagingController(messaging networks.MessagingController) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.Messaging = messaging
}

func (sm *SortingManager) Activate(knownParticipant ma.Multiaddr) {
	log.Println("Activating SortingManager...")
	if knownParticipant == nil {
		log.Println("No known participant. Setting self ID = 0")
		sm.ID = 0
		sm.Self.ID = 0
		sm.ParticipatingNodes[0] = *sm.Self
	} else {
		log.Println("Retrieving participating nodes from known participant:", knownParticipant)
		nodes, err := sm.Messaging.RetrieveParticipatingNodes(sm.Host.Host, knownParticipant, sm.Messaging.GetProtocolID(), sm.Messaging.GetMessageProcessor())
		if err != nil {
			log.Println("Error retrieving participating nodes: ", err)
			return
		}

		sm.ParticipatingNodes = nodes
		sm.ID = utils.MaxKey(nodes) + 1
		sm.Self.ID = sm.ID
		sm.ParticipatingNodes[sm.ID] = *sm.Self

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

		controller, err := networks.DialByMultiaddr(sm.Host.Host, neighbour.Multiaddr, sm.Messaging.GetProtocolID(), sm.Messaging.GetMessageProcessor())
		if err != nil {
			log.Printf("Failed to dial neighbour %s: %v", neighbour.Multiaddr.String(), err)
			continue
		}

		msg := messages.NewAnnounceSelfMessage(sm.ID, sm.Self.Multiaddr)
		go func(c networks.MessagingController, m messages.IMessage) {
			log.Printf("AnnounceSelf: Sending AnnounceSelfMessage to %v\n", neighbour.Multiaddr.String())

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
	log.Println("Participating nodes IDs:", ids)

	for _, id := range ids {
		neighbour := sm.ParticipatingNodes[id]
		log.Printf("Retrieving items from neighbour %d at %s\n", id, neighbour.Multiaddr.String())

		controller, err := networks.DialByMultiaddr(sm.Host.Host, neighbour.Multiaddr, sm.Messaging.GetProtocolID(), sm.Messaging.GetMessageProcessor())
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
				log.Printf("GetAllItems: Unexpected response type from neighbour %d: %T", id, response)
			}
		case <-time.After(3 * time.Second):
			log.Printf("Timeout while waiting for items from neighbour %d", id)
		}

		log.Printf("Finished retrieving items from neighbour %d\n", id)
	}

	return allItems
}

func (sm *SortingManager) ProcessMessage(msg messages.IMessage, controller networks.MessagingController) {
	log.Printf("ProcessMessage: Processing message of type %T from %s\n", msg, controller.GetRemoteAddress())

	switch m := msg.(type) {
	case *messages.CornerItemChangeMessage:
		log.Printf("Received CornerItemChangeMessage: %s", msg)
		sm.ProcessCornerItemChange(*m, controller)
	case *messages.NodesListMessage:
		log.Println("Received NodesListMessage, updating participating nodes...")
		response := messages.NewNodesListResponseMessage(sm.ParticipatingNodes, m.GetTransactionID())
		log.Printf("Sending NodesListResponseMessage with list of %d participating nodes", len(sm.ParticipatingNodes))
		future := controller.SendMessage(response)
		select {
		case <-future:
			log.Println("NodesListResponseMessage sent successfully.")
		case <-time.After(3 * time.Second):
			log.Println("Timeout while sending NodesListResponseMessage.")
		}
	case *messages.NodesListResponseMessage:
		log.Println("Received NodesListResponseMessage, updating participating nodes...")
		sm.mu.Lock()
		defer sm.mu.Unlock()
		for id, neighbour := range m.GetParticipatingNodes() {
			if _, exists := sm.ParticipatingNodes[id]; !exists {
				log.Printf("Adding new neighbour: %d at %s", id, neighbour.Multiaddr.String())
				sm.ParticipatingNodes[id] = neighbour
			} else {
				log.Printf("Updating existing neighbour: %d at %s", id, neighbour.Multiaddr.String())
				sm.ParticipatingNodes[id] = neighbour
			}
		}
	case *messages.AnnounceSelfMessage:
		log.Printf("Received AnnounceSelfMessage from %d: %s", m.ID, m.ListeningAddress)
		neighbour := neighbours.NewNeighbour(m.ListeningAddress, m.ID)
		sm.mu.Lock()
		defer sm.mu.Unlock()
		sm.ParticipatingNodes[m.ID] = *neighbour
		log.Println("Participating nodes: ", sm.ParticipatingNodes)
	case *messages.GetItemsMessage:
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

func (sm *SortingManager) GetItems() []int64 {
	log.Println("Retrieving current items...")
	sm.mu.Lock()
	defer sm.mu.Unlock()
	return sm.Items
}
