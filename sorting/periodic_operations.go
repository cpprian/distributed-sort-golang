package sorting

import (
	"log"
	"time"
)

// startPeriodicCornerItemChanges implements the TimerTask logic.
func (sm *SortingManager) startPeriodicCornerItemChanges() {
	ticker := time.NewTicker(500 * time.Millisecond) // Every 500ms
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
				sm.mu.Unlock()

				sm.sendMessageOnCornerItemChange(leftNeighbour, firstItem)
				sm.sendMessageOnCornerItemChange(rightNeighbour, lastItem)
			}
		case <-sm.ctx.Done(): // Stop when SortingManager context is cancelled
			log.Println("Stopping periodic corner item change messages.")
			return
		}
	}
}

// sendMessageOnCornerItemChange sends messages to the left and right neighbours
func (sm *SortingManager) startPeriodicNodeTimeouts() {
	ticker := time.NewTicker(1 * time.Second) // Every 1 second
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			sm.mu.Lock()
			for nodeID := range sm.LastMessageFromNeighbour {
				if sm.LastMessageFromNeighbour[nodeID]+1500 < time.Now().UnixMilli() {
					log.Printf("Node %d has timed out, processing timeout.", nodeID)
					sm.ProcessNodeTimeout(nodeID)
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
	ticker := time.NewTicker(1 * time.Second) // Every 1 second
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			sm.mu.Lock()
			leftNeighbour := sm.GetLeftNeighbour()
			rightNeighbour := sm.GetRightNeighbour()
			if leftNeighbour != nil {
				sm.sendItemsBackupMessage(leftNeighbour)
			} else {
				sm.sendItemsBackupMessage(rightNeighbour)
			}
			sm.mu.Unlock()
		case <-sm.ctx.Done(): // Stop when SortingManager context is cancelled
			log.Println("Stopping periodic items backup.")
			return
		}
	}
}
