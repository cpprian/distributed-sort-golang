package sorting

import (
	"log"
	"sort"
	"time"

	"github.com/cpprian/distributed-sort-golang/messages"
	"github.com/cpprian/distributed-sort-golang/neighbours"
	"github.com/cpprian/distributed-sort-golang/networks"
	"github.com/cpprian/distributed-sort-golang/utils"
	"github.com/google/uuid"
)

func (sm *SortingManager) AddItem(item int64) {
	log.Println("Adding item: ", item)
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if len(sm.Items) > 0 {
		if item < sm.GetFirstItem() {
			log.Println("Item is less than first item, sending message to left neighbour")
			sm.sendMessageOnCornerItemChange(sm.GetLeftNeighbour(), item)
		}
		if item > sm.GetLastItem() {
			log.Println("Item is greater than last item, sending message to right neighbour")
			sm.sendMessageOnCornerItemChange(sm.GetRightNeighbour(), item)
		}
	} else {
		log.Println("No items in the list, sending messages to both neighbours")
		sm.sendMessageOnCornerItemChange(sm.GetLeftNeighbour(), item)
		sm.sendMessageOnCornerItemChange(sm.GetRightNeighbour(), item)
	}

	sm.Items = append(sm.Items, item)
	sort.Slice(sm.Items, func(i, j int) bool {
		return sm.Items[i] < sm.Items[j]
	})
	log.Printf("Added item: %d. Local items after addition: %v", item, sm.Items)
}

func (sm *SortingManager) sendMessageOnCornerItemChange(neighbour *neighbours.Neighbour, item int64) {
	log.Println("Sending message on corner item change for item:", item)
	if neighbour == nil {
		log.Println("Neighbour is nil, cannot send message on corner item change")
		return
	}

	msg := messages.NewCornerItemChangeMessage(item, sm.ID)

	go func() {
		controller, err := networks.DialByMultiaddr(sm.Host.Host, neighbour.Multiaddr, sm.Messaging.GetProtocolID(), sm.Messaging.GetMessageProcessor())
		if err != nil {
			log.Printf("Failed to dial neighbour %s: %v", neighbour.Multiaddr.String(), err)
			return
		}
		defer controller.Close()

		responseChan := controller.SendMessage(msg)

		select {
		case response := <-responseChan:
			if ItemExchangeMsg, ok := response.(messages.ItemExchangeMessage); ok {
				sm.RespondToItemExchange(ItemExchangeMsg, controller)
			} else {
				log.Printf("Unexpected response type: %T", response)
			}
		case <-time.After(3 * time.Second):
			log.Printf("Timeout while waiting for response from neighbour %s", neighbour.Multiaddr.String())
		}
	}()
}

func (sm *SortingManager) RespondToItemExchange(msg messages.ItemExchangeMessage, controller networks.MessagingController) {
	log.Printf("Responding to item exchange: %v", msg)

	sm.mu.Lock()
	defer sm.mu.Unlock()

	var itemToSend int64
	senderID := msg.GetSenderID()

	if senderID > sm.ID {
		itemToSend = sm.GetLastItem()
		if itemToSend != msg.GetWantedItem() {
			log.Printf("Error: wanted item %d does not match last item %d", msg.GetWantedItem(), itemToSend)
			controller.SendMessage(messages.NewErrorMessage(msg.GetTransactionID()))
			return
		}
		sm.Items[len(sm.Items)-1] = msg.GetOfferedItem()
	} else {
		itemToSend = sm.GetFirstItem()
		if itemToSend != msg.GetWantedItem() {
			log.Printf("Error: wanted item %d does not match first item %d", msg.GetWantedItem(), itemToSend)
			controller.SendMessage(messages.NewErrorMessage(msg.GetTransactionID()))
			return
		}
		sm.Items[0] = msg.GetOfferedItem()
	}

	sort.Slice(sm.Items, func(i, j int) bool {
		return sm.Items[i] < sm.Items[j]
	})

	responseMsg := messages.ItemExchangeMessage{
		BaseMessage: messages.BaseMessage{
			MessageType:   messages.ItemExchange,
			TransactionID: msg.GetTransactionID(),
		},
		OfferedItem: itemToSend,
		WantedItem:  msg.GetWantedItem(),
		SenderID:    sm.ID,
	}

	log.Printf("Sending response with offered item %d for transaction ID %s", itemToSend, msg.GetTransactionID())
	controller.SendMessage(responseMsg)
}

func (sm *SortingManager) OrderItemsExchange(controller networks.MessagingController, offeredItem int64, wantedItem int64, neighbourId int64, transactionId uuid.UUID) {
	log.Printf("Ordering items exchange: offeredItem=%d, wantedItem=%d, neighbourId=%d, transactionId=%s", offeredItem, wantedItem, neighbourId, transactionId)

	sm.mu.Lock()
	removed := false
	for i, v := range sm.Items {
		if v == offeredItem {
			sm.Items = append(sm.Items[:i], sm.Items[i+1:]...)
			removed = true
			break
		}
	}
	sm.mu.Unlock()

	if !removed {
		log.Printf("Offered item %d not found in local items: %v", offeredItem, sm.Items)
		return
	}

	go func() {
		respChan := controller.SendMessage(messages.NewItemExchangeMessageWithID(offeredItem, wantedItem, transactionId, sm.ID))

		select {
		case msg := <-respChan:
			log.Printf("OrderItemsExchange received response: %v", msg)
			reposnse, ok := msg.(messages.ItemExchangeMessage)
			if !ok {
				log.Printf("Unexpected response type: %T", msg)
				sm.mu.Lock()
				sm.Items = append(sm.Items, offeredItem)
				sort.Slice(sm.Items, func(i, j int) bool {
					return sm.Items[i] < sm.Items[j]
				})
				sm.mu.Unlock()
				return
			}

			sm.mu.Lock()
			if neighbourId > sm.ID {
				log.Printf("Neighbour ID is greater than local ID, appending offered item %d to the end of the list", reposnse.GetOfferedItem())
				sm.Items = append(sm.Items, reposnse.GetOfferedItem())
			} else {
				log.Printf("Neighbour ID is less than or equal to local ID, prepending offered item %d to the start of the list", reposnse.GetOfferedItem())
				sm.Items = append([]int64{reposnse.GetOfferedItem()}, sm.Items...)
			}
			sort.Slice(sm.Items, func(i, j int) bool {
				return sm.Items[i] < sm.Items[j]
			})
			log.Printf("Items after exchange: %v", sm.Items)
			sm.mu.Unlock()
		case <-time.After(3 * time.Second):
			log.Printf("Timeout while waiting for response from neighbour %d", neighbourId)
			sm.mu.Lock()
			sm.Items = append(sm.Items, offeredItem)
			sort.Slice(sm.Items, func(i, j int) bool {
				return sm.Items[i] < sm.Items[j]
			})
			log.Printf("Reverted items after timeout: %v", sm.Items)
			sm.mu.Unlock()
		}
	}()
	log.Printf("OrderItemsExchange initiated for offered item %d, wanted item %d, neighbour ID %d, transaction ID %s", offeredItem, wantedItem, neighbourId, transactionId)
}

func (sm *SortingManager) ProcessCornerItemChange(msg messages.CornerItemChangeMessage, controller networks.MessagingController) {
	log.Printf("Processing corner item change: %v\n", msg)

	sm.mu.Lock()
	defer sm.mu.Unlock()

	if msg.SenderID > sm.ID {
		log.Printf("ProcessCornerItemChange: Sender ID %d is greater than local ID %d\n", msg.SenderID, sm.ID)
		if sm.GetLastItem() > msg.Item {
			log.Printf("Exchanging items: local last item %d is greater than incoming item %d\n", sm.GetLastItem(), msg.Item)
			sm.OrderItemsExchange(controller, sm.GetLastItem(), msg.Item, msg.SenderID, msg.GetTransactionID())
		}
	} else {
		log.Printf("ProcessCornerItemChange: Sender ID %d is less than or equal to local ID %d\n", msg.SenderID, sm.ID)
		if sm.GetFirstItem() < msg.Item {
			log.Printf("Exchanging items: local first item %d is less than incoming item %d\n", sm.GetFirstItem(), msg.Item)
			sm.OrderItemsExchange(controller, sm.GetFirstItem(), msg.Item, msg.SenderID, msg.GetTransactionID())
		}
	}

	if !utils.Contains(sm.Items, msg.Item) {
		log.Printf("Item %d not found in local items: %v", msg.Item, sm.Items)
		controller.SendMessage(messages.NewConfirmMessage(msg.GetTransactionID()))
		return
	}
	log.Printf("Item %d found in local items: %v\n", msg.Item, sm.Items)
}
