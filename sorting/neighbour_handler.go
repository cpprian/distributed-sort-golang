package sorting

import (
	"log"
	"sort"

	"github.com/cpprian/distributed-sort-golang/messages"
	"github.com/cpprian/distributed-sort-golang/neighbours"
	"github.com/cpprian/distributed-sort-golang/networks"
)

func (sm *SortingManager) processNodeTimeout(nodeID int64) {
	log.Printf("ProcessNodeTimeout: Sending invalidate messages for node %d due to timeout.", nodeID)

	for _, neighbour := range sm.ParticipatingNodes {
		go func() {
			controller, err := networks.DialByMultiaddr(sm.Host.Host, neighbour.Multiaddr, sm.Messaging.GetProtocolID(), sm.Messaging.GetMessageProcessor())
			if err != nil {
				log.Printf("ProcessNodeTimeout: Failed to dial neighbour %s: %v", neighbour.Multiaddr.String(), err)
				return
			}

			invalidateNodeMessage := messages.NewInvalidateNodeMessage(nodeID)
			responseChan := controller.SendMessage(invalidateNodeMessage)

			response := <-responseChan
			if response == nil {
				log.Printf("ProcessNodeTimeout: Received nil response for invalidateNode from %s", neighbour.Multiaddr.String())
				sm.mu.Lock()
				delete(sm.ParticipatingNodes, neighbour.ID)
				sm.mu.Unlock()
				log.Printf("ProcessNodeTimeout: Node %d has been removed from participating nodes due to timeout.", neighbour.ID)
				controller.Close()
				return
			}

			if errorMessage, ok := response.(*messages.ErrorMessage); ok {
				log.Printf("ProcessNodeTimeout: Error when receiving response for invalidateNode from %s: %v", neighbour.Multiaddr.String(), errorMessage)
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

	var rightNeigbourList []int64
	if backupItems, exists := sm.RightNodeItemsBackup[invalidateNodeMessage.ID]; exists {
		rightNeigbourList := append(rightNeigbourList, backupItems...)
		sm.RightNodeItemsBackup[invalidateNodeMessage.ID] = rightNeigbourList
		sort.Slice(sm.Items, func(i, j int) bool {
			return sm.Items[i] > sm.Items[j]
		})
		log.Printf("ProcessInvalidateMessage: Added backup items from node %d, new items: %v", invalidateNodeMessage.ID, sm.Items)
	}

	responseMsg := messages.NewConfirmMessage(invalidateNodeMessage.TransactionID)
	controller.SendMessage(responseMsg)
	log.Printf("ProcessInvalidateMessage: Sent confirmation for transaction ID %s", invalidateNodeMessage.TransactionID)
}

func (sm *SortingManager) sendItemsBackupMessage(neighbour *neighbours.Neighbour) {
	log.Printf("SendItemsBackupMessage: Sending items backup to neighbour %d at %s", neighbour.ID, neighbour.Multiaddr.String())

	go func() {
		controller, err := networks.DialByMultiaddr(sm.Host.Host, neighbour.Multiaddr, sm.Messaging.GetProtocolID(), sm.Messaging.GetMessageProcessor())
		if err != nil {
			log.Printf("SendItemsBackupMessage: Failed to dial neighbour %s: %v", neighbour.Multiaddr.String(), err)
			return
		}

		itemsBackupMessage := messages.NewItemsBackupMessage(sm.Items, neighbour.GetID())
		responseChan := controller.SendMessage(itemsBackupMessage)

		response := <-responseChan
		if response == nil {
			log.Printf("SendItemsBackupMessage: Received nil response for items backup from %s", neighbour.Multiaddr.String())
			controller.Close()
			return
		}

		if errorMessage, ok := response.(*messages.ErrorMessage); ok {
			log.Printf("SendItemsBackupMessage: Error when receiving response for items backup from %s: %v", neighbour.Multiaddr.String(), errorMessage)
			controller.Close()
		} else {
			log.Printf("SendItemsBackupMessage: Successfully sent items backup to %s", neighbour.Multiaddr.String())
			controller.Close()
		}
	}()
}
