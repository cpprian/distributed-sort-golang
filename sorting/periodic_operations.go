package sorting

import (
	"log"
	"time"
)

// startPeriodicCornerItemChanges implements the TimerTask logic.
func (sm *SortingManager) startPeriodicCornerItemChanges() {
	ticker := time.NewTicker(1500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			sm.mu.Lock()
			if len(sm.Items) > 0 {
				firstItem := sm.GetFirstItem()
				lastItem := sm.GetLastItem()
				leftNeighbour := sm.GetLeftNeighbour()
				rightNeighbour := sm.GetRightNeighbour()

				sm.sendMessageOnCornerItemChange(leftNeighbour, firstItem)
				sm.sendMessageOnCornerItemChange(rightNeighbour, lastItem)
			}
			sm.mu.Unlock()
		case <-sm.ctx.Done(): // Stop when SortingManager context is cancelled
			log.Println("Stopping periodic corner item change messages.")
			return
		}
	}
}

// sendMessageOnCornerItemChange sends messages to the left and right neighbours
func (sm *SortingManager) startPeriodicNodeTimeouts() {
	ticker := time.NewTicker(3000 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			sm.mu.Lock()
			for nodeID := range sm.LastMessageFromNeighbour {
				if sm.LastMessageFromNeighbour[nodeID]+3000 < time.Now().UnixMilli() {
					log.Printf("Node %d has timed out, processing timeout.", nodeID)
					sm.processNodeTimeout(nodeID)
				}
			}
			sm.mu.Unlock()
		case <-sm.ctx.Done(): // Stop when SortingManager context is cancelled
			log.Println("Stopping periodic node timeout checks.")
			return
		}
	}
}

// sendMessageOnCornerItemChange sends messages to the left and right neighbours
func (sm *SortingManager) startPeriodicItemsBackup() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			sm.mu.Lock()
			if sm.ParticipatingNodes == nil {
				log.Println("No participating nodes, skipping items backup.")
				sm.mu.Unlock()
				continue
			}
			leftNeighbour := sm.GetLeftNeighbour()
			rightNeighbour := sm.GetRightNeighbour()
			if leftNeighbour != nil {
				log.Println("Sending items backup to left neighbour. Neighbour ID:", leftNeighbour.ID)
				sm.sendItemsBackupMessage(leftNeighbour)
			} else if rightNeighbour != nil {
				log.Println("No left neighbour found, sending items backup to right neighbour. Neighbour ID:", rightNeighbour.ID)
				sm.sendItemsBackupMessage(rightNeighbour)
			}
			sm.mu.Unlock()
		case <-sm.ctx.Done(): // Stop when SortingManager context is cancelled
			log.Println("Stopping periodic items backup.")
			return
		}
	}
}
