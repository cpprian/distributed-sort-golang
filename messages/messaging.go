package messages

import (
	"errors"
	"sync"

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
	for {
		msg, err := mi.ReadMessage()
		if err != nil {
			// Handle error (e.g., log it, close the stream, etc.)
			return
		}

		mi.mu.Lock()
		future, exists := mi.sentRequests[msg.TransactionID]
		if exists {
			delete(mi.sentRequests, msg.TransactionID)
			future <- msg
			close(future)
		} else {
			go mi.processor(msg, mi)
		}
		mi.mu.Unlock()
	}
}

func (mi *MessagingInitiator) ReadMessage() (BaseMessage, error) {
	// Implement the logic to read a message from the stream.
	// This is a placeholder implementation.
	var msg BaseMessage
	// Read from mi.stream and unmarshal into msg
	return msg, nil // Replace with actual error handling
}

func (mi *MessagingInitiator) SendMessage(msg BaseMessage) <-chan BaseMessage {
	mi.mu.Lock()
	defer mi.mu.Unlock()

	future := make(chan BaseMessage, 1)
	mi.sentRequests[msg.TransactionID] = future

	// Marshal the message and write it to the stream
	// This is a placeholder implementation.
	// Replace with actual marshaling logic.
	// mi.stream.Write(...)

	return future
}

func (mi *MessagingInitiator) Close() {
	mi.mu.Lock()
	defer mi.mu.Unlock()

	for _, future := range mi.sentRequests {
		close(future)
	}
	mi.sentRequests = make(map[uuid.UUID]chan BaseMessage)

	if err := mi.stream.Close(); err != nil {
		// Handle error (e.g., log it)
	}
}

var ErrTimeout = errors.New("messaging timeout")

type MessagingError struct {
	Msg string
}

func (e *MessagingError) Error() string {
	return "MessagingError: " + e.Msg
}

func NewMessagingError(msg string) error {
	return &MessagingError{Msg: msg}
}