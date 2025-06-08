package networking

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/libp2p/go-libp2p/core/network"

	"github.com/cpprian/distributed-sort-golang/messages"
	"github.com/cpprian/distributed-sort-golang/neighbours"

	ma "github.com/multiformats/go-multiaddr"
)

type MessagingController interface {
	SendMessage(msg messages.BaseMessage) <-chan messages.BaseMessage
	RetrieveParticipatingNodes(addr ma.Multiaddr) (map[int64]neighbours.Neighbour, error)
}

type UnknownMessageProcessor func(msg messages.BaseMessage, controller MessagingController)

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
	sentRequests map[uuid.UUID]chan messages.BaseMessage
	mu           sync.Mutex
}

func NewMessagingInitiator(processor UnknownMessageProcessor, stream network.Stream) *MessagingInitiator {
	return &MessagingInitiator{
		stream:       stream,
		processor:    processor,
		sentRequests: make(map[uuid.UUID]chan messages.BaseMessage),
	}
}

func (mi *MessagingInitiator) Run() {
	log.Println("MessagingInitiator started, waiting for messages...")
	decoder := json.NewDecoder(mi.stream)
	for {
		var msg messages.BaseMessage
		if err := decoder.Decode(&msg); err != nil {
			log.Println("Error decoding message: ", err)
			return
		}
		if messages.MessageRegistry[msg.MessageType].RequiresResponse {
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

func (mi *MessagingInitiator) SendMessage(msg messages.BaseMessage) <-chan messages.BaseMessage {
	mi.mu.Lock()
	defer mi.mu.Unlock()

	future := make(chan messages.BaseMessage, 1)
	mi.sentRequests[msg.TransactionID] = future

	encoder := json.NewEncoder(mi.stream)
	if err := encoder.Encode(msg); err != nil {
		log.Println("Error encoding message:", err)
		close(future)
		delete(mi.sentRequests, msg.TransactionID)
	}

	log.Println("Sent message with transaction ID: ", msg.TransactionID)
	log.Println("Message content:", msg)

	if messages.MessageRegistry[msg.MessageType].RequiresResponse {
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
	mi.sentRequests = make(map[uuid.UUID]chan messages.BaseMessage)

	if err := mi.stream.Close(); err != nil {
		log.Println("Error closing stream:", err)
	} else {
		log.Println("Stream closed successfully.")
	}
}

func (mi *MessagingInitiator) RetrieveParticipatingNodes(knownParticipant ma.Multiaddr) (map[int64]neighbours.Neighbour, error) {
	return nil, nil
}
