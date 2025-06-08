package messages

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/libp2p/go-libp2p/core/network"
)

// package pl.barpec12.distributedsort;

// import com.fasterxml.jackson.core.JsonProcessingException;
// import com.fasterxml.jackson.databind.JsonNode;
// import com.fasterxml.jackson.databind.ObjectMapper;
// import io.libp2p.core.ConnectionClosedException;
// import io.libp2p.core.Libp2pException;
// import io.libp2p.core.Stream;
// import io.libp2p.core.multistream.StrictProtocolBinding;
// import io.libp2p.protocol.ProtocolHandler;
// import io.libp2p.protocol.ProtocolMessageHandler;
// import io.netty.buffer.ByteBuf;
// import io.netty.buffer.Unpooled;
// import lombok.Getter;
// import org.jetbrains.annotations.NotNull;
// import pl.barpec12.distributedsort.messages.ErrorMessage;
// import pl.barpec12.distributedsort.messages.Message;

// import java.io.IOException;
// import java.nio.charset.StandardCharsets;
// import java.time.Duration;
// import java.util.*;
// import java.util.concurrent.*;
// import java.util.logging.Level;
// import java.util.logging.Logger;

// import static io.libp2p.etc.types.AsyncExtKt.completedExceptionally;

// interface MessagingController {
//     CompletableFuture<Message> sendMessage(Message message);
// }

// @FunctionalInterface
// interface UnknownMessageProcessor {
//     void process(Message message, MessagingController controller);
// }

// class Messaging extends MessagingBinding {

//     public Messaging(UnknownMessageProcessor processor) {
//         super(new MessagingProtocol(processor));
//     }
// }

// class MessagingBinding extends StrictProtocolBinding<MessagingController> {
//     public MessagingBinding(MessagingProtocol protocol) {
//         super("/p2p/messaging/1.0.0", protocol);
//     }
// }

// class MessagingTimeoutException extends Libp2pException {}

// @Getter
// class MessagingErrorException extends Libp2pException {
//     private final ErrorMessage errorMessage;

//     MessagingErrorException(ErrorMessage errorMessage) {
//         this.errorMessage = errorMessage;
//     }
// }

// class MessagingProtocol extends ProtocolHandler<MessagingController> {
//     private final UnknownMessageProcessor processor;

//     public MessagingProtocol(UnknownMessageProcessor processor) {
//         super(Long.MAX_VALUE, Long.MAX_VALUE);
//         this.processor = processor;
//     }

//     ScheduledExecutorService timeoutScheduler =
//             Executors.newSingleThreadScheduledExecutor();
//     Duration requestTimeout = Duration.ofSeconds(2);

//     @Override
//     protected @NotNull CompletableFuture<MessagingController> onStartInitiator(
//             @NotNull Stream stream) {
//         var handler = new MessagingInitiator(processor);
//         stream.pushHandler(handler);
//         return handler.activeFuture;
//     }

//     @Override
//     protected @NotNull CompletableFuture<MessagingController> onStartResponder(
//             @NotNull Stream stream) {
//         var handler = new MessagingInitiator(processor);
//         stream.pushHandler(handler);
//         return CompletableFuture.completedFuture(handler);
//     }

//     class MessagingInitiator
//             implements ProtocolMessageHandler<ByteBuf>, MessagingController {
//         private final Logger logger = Logger.getLogger(MessagingInitiator.class.getName());
//         private final UnknownMessageProcessor processor;
//         CompletableFuture activeFuture = new CompletableFuture<MessagingController>();

//         final Map<UUID, CompletableFuture<Message>> sentRequests = Collections.synchronizedMap(new HashMap<>());

//         Stream stream;
//         boolean closed = false;

//         public MessagingInitiator(UnknownMessageProcessor processor) {
//             this.processor = processor;
//             logger.setLevel(Level.WARNING);
//         }

//         @Override
//         public CompletableFuture<Message> sendMessage(Message message) {
//             var future = new CompletableFuture<Message>();

//             synchronized (sentRequests) {
//                 if (closed)
//                     return completedExceptionally(new ConnectionClosedException());
//                 var previous = sentRequests.put(message.getTransactionId(), future);
//                 if (previous != null) {previous.complete(null);}
//                 if (message.getMessageType().isRequireResponse()) {
//                     var action = timeoutScheduler.schedule(() -> {
//                         var timeoutedFuture = sentRequests.remove(message.getTransactionId());
//                         if (timeoutedFuture != null) {
//                             logger.warning("Timeout for message: " + message + " " + message.getTransactionId());
//                             timeoutedFuture.completeExceptionally(new TimeoutException());
//                         }
//                     }, requestTimeout.toMillis(), TimeUnit.MILLISECONDS);
//                     future.whenComplete((result, throwable) -> {
//                         action.cancel(true);
//                     });
//                 }
//             }

//             ObjectMapper mapper = new ObjectMapper();
//             byte[] json;
//             try {
//                 json = mapper.writeValueAsBytes(message);
//             } catch (JsonProcessingException e) {
//                 e.printStackTrace();
//                 throw new RuntimeException(e);
//             }
//             ByteBuf buffer = Unpooled.wrappedBuffer(json);

//             stream.writeAndFlush(buffer);
//             logger.info("Sent message " + new String(json, StandardCharsets.UTF_8));

//             return future;
//         }

//         @Override
//         public void onClosed(@NotNull Stream stream) {
//             logger.info("Closing...");
//             closed = true;
//             sentRequests.clear();
//             timeoutScheduler.shutdownNow();
//             activeFuture.complete(new ConnectionClosedException());
//         }

//         @Override
//         public void onActivated(@NotNull Stream stream) {
//             this.stream = stream;
//             activeFuture.complete(this);
//         }

//         @Override
//         public void onMessage(@NotNull Stream stream, ByteBuf msg) {
//             String messageString = new String(msg.toString(StandardCharsets.UTF_8));
//             logger.info(stream.getConnection().localAddress() + " Got message " + messageString);
//             ObjectMapper mapper = new ObjectMapper();
//             try {
//                 JsonNode basicMessage = mapper.readTree(messageString);
//                 String messageTypeString = basicMessage.get("messageType").asText();
//                 logger.info("Got message type " + messageTypeString);
//                 Message message = mapper.readValue(messageString,
//                         Message.MessageType.valueOf(messageTypeString).getClazz());
//                 var future = sentRequests.remove(message.getTransactionId());
//                 if (future != null) {
//                     if(Message.MessageType.ERROR.equals(message.getMessageType())) {
//                         future.completeExceptionally(new MessagingErrorException(
//                                 (ErrorMessage) message));
//                         return;
//                     }
//                     logger.info("Completing future...");
//                     //got response
//                     future.complete(message);
//                 } else {
//                     processor.process(message, this);
//                 }
//             } catch (IOException e) {
//                 logger.severe("Error occurred with message: " + msg.toString(StandardCharsets.UTF_8));
//                 e.printStackTrace();
//                 throw new RuntimeException(e);
//             }

//         }
//     }
// }

type MessagingController interface {
	SendMessage(msg BaseMessage) <-chan BaseMessage
}

type UnknownMessageProcessor func(msg BaseMessage, controller MessagingController)

type MessagingProtocol struct {
	processor UnknownMessageProcessor
}

func NewMessagingProtocol(processor UnknownMessageProcessor) *MessagingProtocol {
	return &MessagingProtocol{processor: processor}
}

func (mp *MessagingProtocol) HandleStream(s network.Stream) {
	controller := NewMessagingInitiator(mp.processor, s)
	go controller.Run()
}

type MessagingInitiator struct {
	stream       network.Stream
	processor    UnknownMessageProcessor
	sentRequests map[uuid.UUID]chan BaseMessage
	mu           sync.Mutex
}

func NewMessagingInitiator(processor UnknownMessageProcessor, stream network.Stream) *MessagingInitiator {
	return &MessagingInitiator{
		stream:       stream,
		processor:    processor,
		sentRequests: make(map[uuid.UUID]chan BaseMessage),
	}
}

func (mi *MessagingInitiator) Run() {
	log.Println("MessagingInitiator started, waiting for messages...")
	decoder := json.NewDecoder(mi.stream)
	for {
		var msg BaseMessage
		if err := decoder.Decode(&msg); err != nil {
			log.Println("Error decoding message: ", err)
			return
		}
		if messageRegistry[msg.MessageType].RequiresResponse {
			mi.mu.Lock()
			log.Println("Received message with transaction ID: ", msg.TransactionID)
			if ch, ok := mi.sentRequests[msg.GetTransactionID()]; ok {
				ch <- msg
				delete(mi.sentRequests, msg.GetTransactionID())
			}
			mi.mu.Unlock()
		} else {
			log.Println("Processing message of type: ", msg.MessageType)
			go mi.processor(msg, mi)
		}
	}
}

func (mi *MessagingInitiator) SendMessage(msg BaseMessage) <-chan BaseMessage {
	mi.mu.Lock()
	defer mi.mu.Unlock()

	future := make(chan BaseMessage, 1)
	mi.sentRequests[msg.TransactionID] = future

	encoder := json.NewEncoder(mi.stream)
	if err := encoder.Encode(msg); err != nil {
		log.Println("Error encoding message:", err)
		close(future)
		delete(mi.sentRequests, msg.TransactionID)
	}

	log.Println("Sent message with transaction ID: ", msg.TransactionID)
	log.Println("Message content:", msg)

	if messageRegistry[msg.MessageType].RequiresResponse {
		go func() {
			select {
			case <-future:
				log.Println("Received response for transaction ID: ", msg.TransactionID)
				mi.mu.Lock()
				delete(mi.sentRequests, msg.TransactionID)
				mi.mu.Unlock()
			case <-time.After(2 * time.Second):
				log.Println("Stream closed before response for transaction ID: ", msg.TransactionID)
				close(future)
				mi.mu.Lock()
				delete(mi.sentRequests, msg.TransactionID)
				mi.mu.Unlock()
			}
		}()
	}

	log.Println("Future created for transaction ID: ", msg.TransactionID)
	return future
}

func (mi *MessagingInitiator) Close() {
	mi.mu.Lock()
	defer mi.mu.Unlock()
	log.Println("Closing MessagingInitiator...")

	for id, future := range mi.sentRequests {
		close(future)
		delete(mi.sentRequests, id)
	}
	mi.sentRequests = make(map[uuid.UUID]chan BaseMessage)

	if err := mi.stream.Close(); err != nil {
		log.Println("Error closing stream:", err)
	} else {
		log.Println("Stream closed successfully.")
	}
}
