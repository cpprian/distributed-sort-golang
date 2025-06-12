package networks

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"

	"github.com/cpprian/distributed-sort-golang/messages"
	"github.com/cpprian/distributed-sort-golang/neighbours"

	ma "github.com/multiformats/go-multiaddr"
)

type MessagingController interface {
	SendMessage(msg messages.IMessage) <-chan messages.IMessage
	Close()
	GetMessageProcessor() UnknownMessageProcessor
	GetProtocolID() protocol.ID
	GetRemoteAddress() ma.Multiaddr
	RetrieveParticipatingNodes(host host.Host, knownParticipant ma.Multiaddr, protocolID protocol.ID, processor UnknownMessageProcessor) (map[int64]neighbours.Neighbour, error)
}

type UnknownMessageProcessor func(msg messages.IMessage, controller MessagingController)

type MessagingProtocol struct {
	processor UnknownMessageProcessor
	initiator *MessagingInitiator
}

func NewMessagingProtocol(processor UnknownMessageProcessor) *MessagingProtocol {
	return &MessagingProtocol{processor: processor}
}

func (mp *MessagingProtocol) HandleStream(s network.Stream) {
	log.Println("Handling new stream for MessagingProtocol...")
	controller := NewMessagingInitiator(mp.processor, s)
	mp.initiator = controller
	log.Println("New stream received, starting MessagingInitiator...")
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

type Envelope struct {
	MessageType   messages.MessageType `json:"messageType"`
	TransactionID uuid.UUID            `json:"transactionId"`
	Raw           json.RawMessage      `json:"-"`
}

func (mi *MessagingInitiator) Run() {
	log.Println("MessagingInitiator started, waiting for messages...")
	decoder := json.NewDecoder(mi.stream)
	for {
		var rawMap map[string]json.RawMessage
		if err := decoder.Decode(&rawMap); err != nil {
			log.Println("Error decoding raw JSON: ", err)
			return
		}
	
		var envelope Envelope
		if err := json.Unmarshal(rawMap["messageType"], &envelope.MessageType); err != nil {
			log.Println("Failed to read messageType:", err)
			continue
		}
		if err := json.Unmarshal(rawMap["transactionId"], &envelope.TransactionID); err != nil {
			log.Println("Failed to read transactionId:", err)
			continue
		}
	
		envelope.Raw, _ = json.Marshal(rawMap)
	
		info, ok := messages.MessageRegistry[envelope.MessageType]
		if !ok {
			log.Println("Run: Unknown message type received:", envelope.MessageType)
			continue
		}
	
		msgPtr := reflect.New(info.GoType).Interface()
		if err := json.Unmarshal(envelope.Raw, msgPtr); err != nil {
			log.Println("Error unmarshalling into typed message: ", err)
			continue
		}
	
		msg, ok := msgPtr.(messages.IMessage)
		if !ok {
			log.Println("Decoded message does not implement IMessage")
			continue
		}
	
		mi.mu.Lock()
		ch, found := mi.sentRequests[msg.GetTransactionID()]
		if found {
			ch <- msg
			log.Println("Run: Received response for transaction ID:", msg.GetTransactionID())
			log.Println("Run: Message content:", msg)
			delete(mi.sentRequests, msg.GetTransactionID())
			mi.mu.Unlock()
			continue
		}
		mi.mu.Unlock()
		log.Println("Run: No future found for transaction ID:", msg.GetTransactionID())
		if mi.processor != nil {
			log.Println("Run: Processing message with custom processor")
			mi.processor(msg, mi)
		} else {
			log.Println("Run: No processor set, ignoring message of type:", envelope.MessageType)
		}
		log.Println("Run: Processed message of type:", envelope.MessageType, "with transaction ID:", msg.GetTransactionID())
		log.Println("Run: Message content:", msg)
		log.Println("Run: Waiting for next message...")		
	}
}

func (mi *MessagingInitiator) SendMessage(msg messages.IMessage) <-chan messages.IMessage {
	mi.mu.Lock()
	future := make(chan messages.IMessage, 1)
	mi.sentRequests[msg.GetTransactionID()] = future
	mi.mu.Unlock()

	encoder := json.NewEncoder(mi.stream)
	log.Println("SendMessage: Writing to stream:", msg.GetTransactionID())
	log.Println("SendMessage: Message content:", msg)
	err := encoder.Encode(msg)
	if err != nil {
		log.Println("SendMessage: Failed to write to stream:", err)
		close(future)
		delete(mi.sentRequests, msg.GetTransactionID())
		return future
	}

	log.Println("SendMessage: Sent message with transaction ID: ", msg.GetTransactionID())
	log.Println("SendMessage: Message content:", msg)

	if messages.MessageRegistry[msg.GetMessageType()].RequiresResponse {
		go func() {
			select {
			case <-future:
				log.Println("SendMessage: Received response for transaction ID: ", msg.GetTransactionID())
				log.Println("SendMessage: Closing future channel for transaction ID: ", msg.GetTransactionID())
				mi.mu.Lock()
				delete(mi.sentRequests, msg.GetTransactionID())
				mi.mu.Unlock()
			case <-time.After(2 * time.Second):
				log.Println("SendMessage: Stream closed before response for transaction ID: ", msg.GetTransactionID())
				close(future)
				mi.mu.Lock()
				delete(mi.sentRequests, msg.GetTransactionID())
				mi.mu.Unlock()
			}
		}()
	}

	log.Println("SendMessage: Future created for transaction ID: ", msg.GetTransactionID())
	return future
}

func (mi *MessagingInitiator) Close() {
	mi.mu.Lock()
	defer mi.mu.Unlock()
	log.Println("Close: Closing MessagingInitiator...")

	for id, future := range mi.sentRequests {
		close(future)
		delete(mi.sentRequests, id)
	}
	mi.sentRequests = make(map[uuid.UUID]chan messages.IMessage)

	if err := mi.stream.Close(); err != nil {
		log.Println("Close: Error closing stream:", err)
	} else {
		log.Println("Close: Stream closed successfully.")
	}
}

func (mi *MessagingInitiator) RetrieveParticipatingNodes(
	host host.Host,
	knownParticipant ma.Multiaddr,
	protocolID protocol.ID,
	processor UnknownMessageProcessor,
) (map[int64]neighbours.Neighbour, error) {
	log.Println("RetrieveParticipatingNodes: Retrieving participating nodes...")
	controller, err := DialByMultiaddr(host, knownParticipant, protocolID, processor)
	if err != nil {
		return nil, fmt.Errorf("RetrieveParticipatingNodes: failed to dial by multiaddr: %w", err)
	}

	msg := messages.NewNodesListMessage()
	future := controller.SendMessage(msg)

	log.Println("RetrieveParticipatingNodes: Sent NodesListMessage, waiting for response...")
	select {
	case responseMsg := <-future:
		response, ok := responseMsg.(*messages.NodesListResponseMessage)
		if !ok {
			return nil, fmt.Errorf("RetrieveParticipatingNodes: unexpected message type: %T", responseMsg)
		}
		log.Println("RetrieveParticipatingNodes: Received nodes list response:", response.GetParticipatingNodes())
		return response.GetParticipatingNodes(), nil
	case <-time.After(3 * time.Second):
		log.Println("RetrieveParticipatingNodes: Timeout waiting for nodes list response")
		return nil, fmt.Errorf("RetrieveParticipatingNodes: timeout waiting for nodes list response")
	}
}

func (mi *MessagingInitiator) GetMessageProcessor() UnknownMessageProcessor {
	return mi.processor
}

func (mi *MessagingInitiator) GetProtocolID() protocol.ID {
	return ProtocolID
}

func (mi *MessagingInitiator) GetRemoteAddress() ma.Multiaddr {
	if mi.stream == nil {
		return nil
	}

	addrInfo := peer.AddrInfo{
		ID:    mi.stream.Conn().RemotePeer(),
		Addrs: []ma.Multiaddr{mi.stream.Conn().RemoteMultiaddr()},
	}
	
	fullAddrs, err := peer.AddrInfoToP2pAddrs(&addrInfo)
	if err != nil || len(fullAddrs) == 0 {
		log.Fatalf("MI_GetRemoteAddress: Failed to convert to full multiaddr: %v", err)
	}
	
	return fullAddrs[0]
}

func (mp *MessagingProtocol) SetInitiator(initiator *MessagingInitiator) {
	mp.initiator = initiator
}

func (mp *MessagingProtocol) SendMessage(msg messages.IMessage) <-chan messages.IMessage {
	if mp.initiator == nil {
		return nil
	}
	return mp.initiator.SendMessage(msg)
}

func (mp *MessagingProtocol) Close() {
	if mp.initiator != nil {
		mp.initiator.Close()
	}
}

func (mp *MessagingProtocol) GetMessageProcessor() UnknownMessageProcessor {
	return mp.processor
}

func (mp *MessagingProtocol) GetProtocolID() protocol.ID {
	return ProtocolID
}

func (mp *MessagingProtocol) GetRemoteAddress() ma.Multiaddr {
	if mp.initiator != nil {
		return mp.initiator.GetRemoteAddress()
	}
	return nil
}

func (mp *MessagingProtocol) RetrieveParticipatingNodes(
	host host.Host,
	knownParticipant ma.Multiaddr,
	protocolID protocol.ID,
	processor UnknownMessageProcessor,
) (map[int64]neighbours.Neighbour, error) {
	initiator, err := mp.Dial(host, knownParticipant)
	if err != nil {
		return nil, fmt.Errorf("RetrieveParticipatingNodes: failed to dial known participant: %w", err)
	}

	msg := messages.NewNodesListMessage()
	future := initiator.SendMessage(msg)

	select {
	case responseMsg := <-future:
		response, ok := responseMsg.(*messages.NodesListResponseMessage)
		if !ok {
			return nil, fmt.Errorf("RetrieveParticipatingNodes: unexpected message type: %T", responseMsg)
		}
		return response.GetParticipatingNodes(), nil
	case <-time.After(3 * time.Second):
		return nil, fmt.Errorf("RetrieveParticipatingNodes: timeout waiting for nodes list response")
	}
}

func (mp *MessagingProtocol) Dial(host host.Host, addr ma.Multiaddr) (*MessagingInitiator, error) {
	peerInfo, err := peer.AddrInfoFromP2pAddr(addr)
	if err != nil {
		return nil, fmt.Errorf("Dial: invalid multiaddr (can't parse peer info): %w", err)
	}

	log.Printf("Dial: Dialing peer %s at %v\n", peerInfo.ID, peerInfo.Addrs)

	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()

	if err := host.Connect(ctx, *peerInfo); err != nil {
		return nil, fmt.Errorf("Dial: failed to connect to peer %s: %w", peerInfo.ID, err)
	}

	var stream network.Stream
	for i := 1; i <= 3; i++ {
		time.Sleep(time.Duration(i) * time.Second)

		supported, err := host.Peerstore().SupportsProtocols(peerInfo.ID, mp.GetProtocolID())
		if err != nil {
			log.Printf("Dial: [Attempt %d] Could not determine supported protocols yet: %v", i, err)
		} else {
			log.Printf("Dial: [Attempt %d] Peer %s supports: %v", i, peerInfo.ID, supported)
		}

		stream, err = host.NewStream(context.Background(), peerInfo.ID, mp.GetProtocolID())
		if err == nil {
			log.Printf("Dial: Successfully opened stream on attempt %d", i)
			break
		} else {
			log.Printf("Dial: [Attempt %d] Failed to open stream: %v", i, err)
		}
	}

	if stream == nil {
		return nil, fmt.Errorf("Dial: failed to open stream after retries")
	}

	initiator := NewMessagingInitiator(mp.processor, stream)
	go initiator.Run()

	log.Printf("Dial: Successfully dialed peer %s with protocol %s", peerInfo.ID, mp.GetProtocolID())
	return initiator, nil
}
