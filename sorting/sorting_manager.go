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

// import io.libp2p.core.Host;
// import io.libp2p.core.multiformats.Multiaddr;
// import lombok.Getter;
// import lombok.NoArgsConstructor;
// import lombok.Setter;

// import java.util.*;
// import java.util.concurrent.CompletableFuture;
// import java.util.concurrent.ExecutionException;
// import java.util.logging.Level;
// import java.util.logging.Logger;

// @Setter
// @Getter
// @NoArgsConstructor
// public class SortingManager<T extends Comparable<T>> {
//     private Logger logger = Logger.getLogger(this.getClass().getName());
//     private Long id;
//     private Map<Long, Neighbour> participatingNodes = new HashMap<>();
//     private List<T> items = new ArrayList<>();
//     private Messaging messaging;
//     private Host host;
//     private Neighbour self;

//     public void activate(Multiaddr knownParticipant) {
//         logger.setLevel(Level.WARNING);
//         try {
//             host.start().get();
//         } catch (InterruptedException e) {
//             throw new RuntimeException(e);
//         } catch (ExecutionException e) {
//             throw new RuntimeException(e);
//         }
//         if (knownParticipant == null) {
//             id = 0L;
//             self = new Neighbour(host.listenAddresses().get(0), id);
//             participatingNodes.put(id, self);
//         } else {
// //            retrieveParticipatingNodes(knownParticipant).whenComplete((map, throwable) -> {
// //                participatingNodes = map;
//             try {
//                 participatingNodes = retrieveParticipatingNodes(knownParticipant).get();
//             } catch (InterruptedException e) {
//                 throw new RuntimeException(e);
//             } catch (ExecutionException e) {
//                 throw new RuntimeException(e);
//             }
//             Long maxId =
//                         participatingNodes.keySet().stream().max(Long::compare)
//                                 .get();
//                 id = maxId + 1;
//                 self = new Neighbour(host.listenAddresses().get(0), id);
//                 participatingNodes.put(id, self);
//                 announceSelf();
// //            });

//         }

//         new Timer().schedule(new TimerTask() {
//             @Override
//             public void run() {
//                 if(items.size() > 0) {
//                     sendMessageOnCornerItemChange(getLeftNeighbour(),
//                             getFirstItem());
//                     sendMessageOnCornerItemChange(getRightNeighbour(),
//                             getLastItem());
//                 }
//             }
//         }, 0L, 500L);

//     }

//     public void add(T item) {
//         if (item == null) {
//             return;
//         }
//         if (!items.isEmpty()) {
//             //check if first item will change
//             if (item.compareTo(getFirstItem()) < 0) {
//                 sendMessageOnCornerItemChange(getLeftNeighbour(), item);
//             }

//             //check if last item will change
//             if (item.compareTo(getLastItem()) > 0) {
//                 sendMessageOnCornerItemChange(getRightNeighbour(), item);
//             }
//         } else {
//             sendMessageOnCornerItemChange(getLeftNeighbour(), item);
//             sendMessageOnCornerItemChange(getRightNeighbour(), item);
//         }
//         synchronized (items) {
//             items.add(item);
//             items.sort(Comparable::compareTo);
//         }
//     }

//     public void remove(T item) {
//         if (item == null) {
//             return;
//         }
//         items.remove(item);
//     }

//     public T getFirstItem() {
//         return items.get(0);
//     }

//     public T getLastItem() {
//         return items.get(items.size() - 1);
//     }

//     public Optional<Neighbour> getRightNeighbour() {
//         var neighbour = participatingNodes.entrySet().stream()
//                                           .filter(entry -> entry.getKey() >
//                                                   id)
//                                           .min(Map.Entry.comparingByKey())
//                                           .map(Map.Entry::getValue);
//         return neighbour;
//     }

//     public Optional<Neighbour> getLeftNeighbour() {
//         var neighbour = participatingNodes.entrySet().stream()
//                                           .filter(entry -> entry.getKey() <
//                                                   id)
//                                           .max(Map.Entry.comparingByKey())
//                                           .map(Map.Entry::getValue);
//         return neighbour;
//     }

//     public void sendMessageOnCornerItemChange(
//             Optional<Neighbour> neighbour,
//             T item) {
//         if (neighbour.isEmpty()) {
//             return;
//         }
//         CornerItemChangeMessage message =
//                 new CornerItemChangeMessage(item, id);
//         messaging.dial(host, neighbour.get().getMultiaddr())
//                  .getController()
//                  .whenCompleteAsync((controller, throwable) -> controller.sendMessage(message).whenCompleteAsync((response, throwable1) -> respondToItemsExchange(
//                          (ItemExchangeMessage<T>) response, controller)));
//     }

//     public void processCornerItemChange (
//             CornerItemChangeMessage<T> message, MessagingController controller) {
//         boolean sent = false;
//         if (message.getSenderId() > id) {
//             if (getLastItem().compareTo(message.getItem()) > 0) {
//                 orderItemsExchange(
//                         controller,
//                         getLastItem(),
//                         message.getItem(),
//                         message.getSenderId(),
//                         message.getTransactionId());
//                 sent = true;
//             }
//         } else {
//             if (getFirstItem().compareTo(message.getItem()) < 0) {
//                 orderItemsExchange(
//                         controller,
//                         getFirstItem(),
//                         message.getItem(),
//                         message.getSenderId(),
//                         message.getTransactionId());
//                 sent = true;
//             }
//         }
//         if (!sent) {
//             controller.sendMessage(new ConfirmMessage(message.getTransactionId()));
//         }
//     }

//     public void orderItemsExchange(MessagingController controller, T offeredItem, T wantedItem, long neighbourId, UUID transactionId) {
//         CompletableFuture<Message> future =
//                 controller.sendMessage(
//                         new ItemExchangeMessage<>(offeredItem, wantedItem, transactionId, id));
//         if(!items.remove(offeredItem)) {
//             logger.severe("Could not remove item " + offeredItem + ". It was not present!");
//         }
//         future.whenCompleteAsync((message, throwable) -> {
//             ItemExchangeMessage<T> response =
//                     (ItemExchangeMessage<T>) message;

//             if (neighbourId > id) {
//                 items.add(items.size(), response.getOfferedItem());
//             } else {
//                 items.add(0, response.getOfferedItem());
//             }
//             items.sort(Comparable::compareTo);
//         }).exceptionally(throwable -> {
//             logger.info("Reverting...");
//             items.add(offeredItem);
//             items.sort(Comparable::compareTo);
//             return null;
//         });

//     }

//     public void respondToItemsExchange(ItemExchangeMessage<T> message,
//                                        MessagingController controller) {
//         T itemToSend;
//         if (message.getSenderId() > id) {
//             itemToSend = items.get(items.size() - 1);
//             if (!itemToSend.equals(message.getWantedItem())) {
//                 controller.sendMessage(new ErrorMessage(message.getTransactionId()));
//                 return;
//             }
//             items.set(items.size() - 1, message.getOfferedItem());
//         } else {
//             itemToSend = items.get(0);
//             if(!itemToSend.equals(message.getWantedItem())) {
//                 controller.sendMessage(new ErrorMessage(message.getTransactionId()));
//                 return;
//             }
//             items.set(0, message.getOfferedItem());
//         }
//         items.sort(Comparable::compareTo);

//         controller.sendMessage(new ItemExchangeMessage<>(itemToSend,
//                 message.getWantedItem(),
//                 message.getTransactionId(), id));
//         //TODO possible improvement - response might get lost in transit
//     }

//     public void processMessage(Message message,
//                                MessagingController controller) {
//         if (message instanceof CornerItemChangeMessage changeMessage) {
//             processCornerItemChange(changeMessage, controller);
//         } else if (message instanceof NodesListMessage nodesListMessage) {
//             NodesListResponseMessage response =
//                     new NodesListResponseMessage(participatingNodes,
//                             nodesListMessage.getTransactionId());
//             controller.sendMessage(response);
//         } else if (message instanceof AnnounceSelfMessage announceSelfMessage) {
//             participatingNodes.put(announceSelfMessage.getId(),
//                     new Neighbour(
//                             announceSelfMessage.getListeningAddress(),
//                             announceSelfMessage.getId()));
//         } else if (message instanceof GetItemsMessage<?> getItemsMessage) {
//             GetItemsMessage response = new GetItemsMessage(items, getItemsMessage.getTransactionId());
//             controller.sendMessage(response);
//         }

//     }

//     public CompletableFuture<Map<Long, Neighbour>> retrieveParticipatingNodes(
//             Multiaddr knownParticipant) {
//         var response = messaging.dial(host, knownParticipant).getController().thenCompose(controller -> {
//             CompletableFuture<Message> future =
//                     controller.sendMessage(new NodesListMessage());
//             return future.thenApplyAsync(o -> ((NodesListResponseMessage) o).getParticipatingNodes());
//         });

//         return response;
//     }

//     public void announceSelf() {
//         for (Neighbour neighbour : participatingNodes.values()) {
//             if (neighbour.getId() != id){
//                 messaging.dial(host, neighbour.getMultiaddr())
//                          .getController().thenApplyAsync(
//                                  controller -> controller.sendMessage(
//                                          new AnnounceSelfMessage(id,
//                                                  host.listenAddresses().get(0))));
//             }
//         }
//     }

//     public List<T> getAllItems() {
//         List<T> items = new ArrayList<>();
//         for (Neighbour neighbour : participatingNodes.entrySet().stream()
//                                                      .sorted(Map.Entry.comparingByKey())
//                                                      .map(
//                                                              Map.Entry::getValue)
//                                                      .toList()) {
//             try {
//                 var itemsFromNode = messaging.dial(host, neighbour.getMultiaddr())
//                          .getController().thenCompose(controller -> {
//                              CompletableFuture<Message> future =
//                                      controller.sendMessage(
//                                              new GetItemsMessage<T>());
//                              return future.thenApplyAsync(message -> ((GetItemsMessage)message).getItems());
//                          });
//                 items.addAll(itemsFromNode.get());

//             } catch (InterruptedException e) {
//                 throw new RuntimeException(e);
//             } catch (ExecutionException e) {
//                 throw new RuntimeException(e);
//             }
//         }
//         return items;
//     }

// }

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

func (sm *SortingManager) AddItem(item int64) {
	log.Println("Adding item: ", item)
	// TODO: implement AddItem here
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

func (sm *SortingManager) GetNeighbour(isRight bool) *neighbours.Neighbour {
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
	return sm.GetNeighbour(true)
}

func (sm *SortingManager) GetLeftNeighbour() *neighbours.Neighbour {
	return sm.GetNeighbour(false)
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