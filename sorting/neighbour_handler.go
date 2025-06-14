package sorting

import (
	"log"
	"sort"

	"github.com/cpprian/distributed-sort-golang/messages"
	"github.com/cpprian/distributed-sort-golang/neighbours"
	"github.com/cpprian/distributed-sort-golang/networks"
)

func (sm *SortingManager) ProcessNodeTimeout(nodeID int64) {
	log.Printf("ProcessNodeTimeout: Sending invalidate messages for node %d due to timeout.", nodeID)
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for _, neighbour := range sm.ParticipatingNodes {
		controller, err := networks.DialByMultiaddr(sm.Host.Host, neighbour.Multiaddr, sm.Messaging.GetProtocolID(), sm.Messaging.GetMessageProcessor())
		if err != nil {
			log.Printf("ProcessNodeTimeout: Failed to dial neighbour %s: %v", neighbour.Multiaddr.String(), err)
			continue
		}

		invalidateNodeMessage := messages.NewInvalidateNodeMessage(nodeID)
		responseChan := controller.SendMessage(invalidateNodeMessage)

		go func() {
			response := <-responseChan
			if response.(*messages.ErrorMessage) != nil {
				log.Printf("ProcessNodeTimeout: Error when receiving response for invalidateNode from %s", neighbour.Multiaddr.String())
				sm.mu.Lock()
				delete(sm.ParticipatingNodes, neighbour.ID)
				sm.mu.Unlock()
			} else {
				log.Printf("ProcessNodeTimeout: Successfully sent invalidateNode to %s", neighbour.Multiaddr.String())
				controller.Close()
			}
		}()
	}
}

func (sm *SortingManager) processInvalidateMessage(invalidateNodeMessage *messages.InvalidateNodeMessage, controller networks.MessagingController) {
	log.Printf("ProcessInvalidateMessage: Invalidating node %d due to timeout.", invalidateNodeMessage.ID)
	sm.mu.Lock()
	defer sm.mu.Unlock()

	delete(sm.ParticipatingNodes, invalidateNodeMessage.GetID())
	delete(sm.LastMessageFromNeighbour, invalidateNodeMessage.GetID())

	if backupItems, exists := sm.RightNodeItemsBackup[invalidateNodeMessage.ID]; exists {
		sm.Items = append(sm.Items, backupItems...)
		sort.Slice(sm.Items, func(i, j int) bool {
			return sm.Items[i] > sm.Items[j]
		})
		delete(sm.RightNodeItemsBackup, invalidateNodeMessage.ID)
		log.Printf("ProcessInvalidateMessage: Added backup items from node %d, new items: %v", invalidateNodeMessage.ID, sm.Items)
	}

	responseMsg := messages.NewConfirmMessage(invalidateNodeMessage.TransactionID)
	controller.SendMessage(responseMsg)
	log.Printf("ProcessInvalidateMessage: Sent confirmation for transaction ID %s", invalidateNodeMessage.TransactionID)
}

func (sm *SortingManager) sendItemsBackupMessage(neighbour *neighbours.Neighbour) {
	log.Printf("SendItemsBackupMessage: Sending items backup to neighbour %d at %s", neighbour.ID, neighbour.Multiaddr.String())
	sm.mu.Lock()
	defer sm.mu.Unlock()

	controller, err := networks.DialByMultiaddr(sm.Host.Host, neighbour.Multiaddr, sm.Messaging.GetProtocolID(), sm.Messaging.GetMessageProcessor())
	if err != nil {
		log.Printf("SendItemsBackupMessage: Failed to dial neighbour %s: %v", neighbour.Multiaddr.String(), err)
		return
	}

	itemsBackupMessage := messages.NewItemsBackupMessage(sm.Items, neighbour.GetID())
	responseChan := controller.SendMessage(itemsBackupMessage)

	go func() {
		response := <-responseChan
		if response.(*messages.ErrorMessage) != nil {
			log.Printf("SendItemsBackupMessage: Error when receiving response for items backup from %s", neighbour.Multiaddr.String())
			controller.Close()
		} else {
			log.Printf("SendItemsBackupMessage: Successfully sent items backup to %s", neighbour.Multiaddr.String())
			controller.Close()
		}
	}()
}
