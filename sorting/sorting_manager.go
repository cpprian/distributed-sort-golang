package sorting

import (
	"log"
	"sync"

	"github.com/cpprian/distributed-sort-golang/neighbours"
	"github.com/cpprian/distributed-sort-golang/networking"
	host "github.com/libp2p/go-libp2p/core/host"
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
	ParticipatingNodes map[int64]string
	Host               host.Host
	Neighbour          neighbours.Neighbour
	Messaging          networking.MessagingController
	mu                 sync.Mutex
}

func NewSortingManager() *SortingManager {
	id := int64(0)

	log.Println("Creating SortingManager...")

	return &SortingManager{
		ID:                 id,
		Items:              []int64{},
		ParticipatingNodes: make(map[int64]string),
		Host:               nil,
		Messaging:          nil,
		mu:                 sync.Mutex{},
	}
}
