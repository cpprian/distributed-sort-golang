package networking

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/protocol"

	"github.com/cpprian/distributed-sort-golang/messages"
	"github.com/cpprian/distributed-sort-golang/neighbours"

	ma "github.com/multiformats/go-multiaddr"
)

type MessagingController interface {
	SendMessage(msg messages.IMessage) <-chan messages.IMessage
	Close()
	GetUnknownMessageProcessor() UnknownMessageProcessor
	GetProtocolID() protocol.ID
	RetrieveParticipatingNodes(host host.Host, knownParticipant ma.Multiaddr, protocolID protocol.ID, processor UnknownMessageProcessor) (map[int64]neighbours.Neighbour, error)
}

type UnknownMessageProcessor func(msg messages.IMessage, controller MessagingController)

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
	sentRequests map[uuid.UUID]chan messages.IMessage
	mu           sync.Mutex
}

func NewMessagingInitiator(processor UnknownMessageProcessor, stream network.Stream) *MessagingInitiator {
	return &MessagingInitiator{
		stream:       stream,
		processor:    processor,
		sentRequests: make(map[uuid.UUID]chan messages.IMessage),
	}
}

func (mi *MessagingInitiator) Run() {
	log.Println("MessagingInitiator started, waiting for messages...")
	decoder := json.NewDecoder(mi.stream)
	for {
		var msg messages.IMessage
		if err := decoder.Decode(&msg); err != nil {
			log.Println("Error decoding message: ", err)
			return
		}
		if messages.MessageRegistry[msg.GetMessageType()].RequiresResponse {
			mi.mu.Lock()
			log.Println("Received message with transaction ID: ", msg.GetTransactionID())
			if ch, ok := mi.sentRequests[msg.GetTransactionID()]; ok {
				ch <- msg
				delete(mi.sentRequests, msg.GetTransactionID())
			}
			mi.mu.Unlock()
		} else {
			log.Println("Processing message of type: ", msg.GetMessageType())
			go mi.processor(msg, mi)
		}
	}
}

func (mi *MessagingInitiator) SendMessage(msg messages.IMessage) <-chan messages.IMessage {
	mi.mu.Lock()
	defer mi.mu.Unlock()

	future := make(chan messages.IMessage, 1)
	mi.sentRequests[msg.GetTransactionID()] = future

	encoder := json.NewEncoder(mi.stream)
	if err := encoder.Encode(msg); err != nil {
		log.Println("Error encoding message:", err)
		close(future)
		delete(mi.sentRequests, msg.GetTransactionID())
	}

	log.Println("Sent message with transaction ID: ", msg.GetTransactionID())
	log.Println("Message content:", msg)

	if messages.MessageRegistry[msg.GetMessageType()].RequiresResponse {
		go func() {
			select {
			case <-future:
				log.Println("Received response for transaction ID: ", msg.GetTransactionID())
				mi.mu.Lock()
				delete(mi.sentRequests, msg.GetTransactionID())
				mi.mu.Unlock()
			case <-time.After(2 * time.Second):
				log.Println("Stream closed before response for transaction ID: ", msg.GetTransactionID())
				close(future)
				mi.mu.Lock()
				delete(mi.sentRequests, msg.GetTransactionID())
				mi.mu.Unlock()
			}
		}()
	}

	log.Println("Future created for transaction ID: ", msg.GetTransactionID())
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
	mi.sentRequests = make(map[uuid.UUID]chan messages.IMessage)

	if err := mi.stream.Close(); err != nil {
		log.Println("Error closing stream:", err)
	} else {
		log.Println("Stream closed successfully.")
	}
}

func (mi *MessagingInitiator) RetrieveParticipatingNodes(
	host host.Host,
	knownParticipant ma.Multiaddr,
	protocolID protocol.ID,
	processor UnknownMessageProcessor,
) (map[int64]neighbours.Neighbour, error) {
	log.Println("Retrieving participating nodes...")
	controller, err := DialByMultiaddr(host, knownParticipant, protocolID, processor)
	if err != nil {
		return nil, fmt.Errorf("failed to dial by multiaddr: %w", err)
	}

	msg := messages.NewNodesListMessage()
	future := controller.SendMessage(msg)

	select {
	case responseMsg := <-future:
		response, ok := responseMsg.(messages.NodesListResponseMessage)
		if !ok {
			return nil, fmt.Errorf("unexpected message type: %T", responseMsg)
		}
		return response.GetParticipatingNodes(), nil
	case <-time.After(3 * time.Second):
		log.Println("Timeout waiting for nodes list response")
		return nil, fmt.Errorf("timeout waiting for nodes list response")
	}
}

func (mi *MessagingInitiator) GetUnknownMessageProcessor() UnknownMessageProcessor {
	return mi.processor
}

func (mi *MessagingInitiator) GetProtocolID() protocol.ID {
	return mi.stream.Protocol()
}
