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
		if sm.Self.GetID() == neighbour.ID {
			sm.handleNeighbourTimeouts(nodeID)
			continue
		}

		controller, err := networks.DialByMultiaddr(sm.Host.Host, neighbour.Multiaddr, sm.Messaging.GetProtocolID(), sm.Messaging.GetMessageProcessor())
		if err != nil {
			log.Printf("ProcessNodeTimeout: Failed to dial neighbour %s: %v", neighbour.Multiaddr.String(), err)
			sm.handleNeighbourTimeouts(nodeID)
			return
		}

		invalidateNodeMessage := messages.NewInvalidateNodeMessage(nodeID)
		responseChan := controller.SendMessage(invalidateNodeMessage)

		if response := <-responseChan; response != nil {
			log.Printf("ProcessNodeTimeout: Error when receiving response for invalidateNode from %s: %v", neighbour.Multiaddr.String(), response)
			sm.handleNeighbourTimeouts(nodeID)
		} else {
			log.Printf("ProcessNodeTimeout: Successfully sent invalidateNode to %s", neighbour.Multiaddr.String())
			controller.Close()
		}
	}
}

func (sm *SortingManager) processInvalidateMessage(invalidateNodeMessage *messages.InvalidateNodeMessage, controller networks.MessagingController) {
	log.Printf("ProcessInvalidateMessage: Invalidating node %d due to timeout.", invalidateNodeMessage.ID)
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.handleNeighbourTimeouts(invalidateNodeMessage.GetID())

	responseMsg := messages.NewConfirmMessage(invalidateNodeMessage.TransactionID)
	controller.SendMessage(responseMsg)
	log.Printf("ProcessInvalidateMessage: Sent confirmation for transaction ID %s", invalidateNodeMessage.TransactionID)
}

func (sm *SortingManager) handleNeighbourTimeouts(nodeID int64) {
	delete(sm.ParticipatingNodes, nodeID)
	delete(sm.LastMessageFromNeighbour, nodeID)

	log.Println("handleNeighbourTimeouts: Backup items: ", sm.RightNodeItemsBackup[nodeID], " from node ", nodeID)
	if backupItems, exists := sm.RightNodeItemsBackup[nodeID]; exists {
		sm.Items = append(sm.Items, backupItems...)
		delete(sm.RightNodeItemsBackup, nodeID)
		sort.Slice(sm.Items, func(i, j int) bool {
			return sm.Items[i] > sm.Items[j]
		})
		log.Printf("handleNeighbourTimeouts: Added backup items from node %d, new items: %v", nodeID, sm.Items)
	}
}

func (sm *SortingManager) sendItemsBackupMessage(neighbour *neighbours.Neighbour) {
	log.Printf("SendItemsBackupMessage: Sending items backup to neighbour %d at %s", neighbour.ID, neighbour.Multiaddr.String())

	controller, err := networks.DialByMultiaddr(sm.Host.Host, neighbour.Multiaddr, sm.Messaging.GetProtocolID(), sm.Messaging.GetMessageProcessor())
	if err != nil {
		log.Printf("SendItemsBackupMessage: Failed to dial neighbour %s: %v", neighbour.Multiaddr.String(), err)
		return
	}

	itemsBackupMessage := messages.NewItemsBackupMessage(sm.Items, sm.Self.GetID())
	log.Println("SendItemsBackupMessage: Items backup message created with items:", sm.Items)
	controller.SendMessage(itemsBackupMessage)
}
