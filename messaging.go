package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	libp2pnetwork "github.com/libp2p/go-libp2p/core/network"
)

// MessageType represents the type of a message.
type MessageType string

const (
	ItemExchange      MessageType = "ITEM_EXCHANGE"
	CornerItemChange  MessageType = "CORNER_ITEM_CHANGE"
	NodesList         MessageType = "NODES_LIST"
	NodesListResponse MessageType = "NODES_LIST_RESPONSE"
	AnnounceSelf      MessageType = "ANNOUNCE_SELF"
	GetItems          MessageType = "GET_ITEMS"
	Confirm           MessageType = "CONFIRM"
	Error             MessageType = "ERROR"
)

// Message represents a generic message structure.
type Message struct {
	MessageType   MessageType `json:"messageType"`
	TransactionID uuid.UUID   `json:"transactionId"`
	// Additional fields specific to each message type can be embedded or extended.
}

// MessagingController defines the interface for sending messages.
type MessagingController interface {
	SendMessage(ctx context.Context, msg *Message) (*Message, error)
}

// UnknownMessageProcessor defines a function to process unknown messages.
type UnknownMessageProcessor func(msg *Message, controller MessagingController)

// MessagingProtocol handles the libp2p protocol logic.
type MessagingProtocol struct {
	processor        UnknownMessageProcessor
	sentRequests     sync.Map // map[uuid.UUID]chan *Message
	requestTimeout   time.Duration
	timeoutScheduler *time.Timer
}

// NewMessagingProtocol creates a new MessagingProtocol instance.
func NewMessagingProtocol(processor UnknownMessageProcessor) *MessagingProtocol {
	return &MessagingProtocol{
		processor:      processor,
		requestTimeout: 2 * time.Second,
	}
}

// HandleStream handles incoming libp2p streams.
func (mp *MessagingProtocol) HandleStream(stream libp2pnetwork.Stream) {
	go mp.handleIncomingStream(stream)
}

// SendMessage sends a message over the stream and waits for a response.
func (mp *MessagingProtocol) SendMessage(ctx context.Context, stream libp2pnetwork.Stream, msg *Message) (*Message, error) {
	if msg.TransactionID == uuid.Nil {
		msg.TransactionID = uuid.New()
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}

	// Write the message with a newline delimiter.
	_, err = stream.Write(append(data, '\n'))
	if err != nil {
		return nil, err
	}

	// Prepare a channel to receive the response.
	responseCh := make(chan *Message, 1)
	mp.sentRequests.Store(msg.TransactionID, responseCh)
	defer mp.sentRequests.Delete(msg.TransactionID)

	// Wait for the response or timeout.
	select {
	case resp := <-responseCh:
		return resp, nil
	case <-time.After(mp.requestTimeout):
		return nil, errors.New("request timed out")
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// handleIncomingStream processes messages from the stream.
func (mp *MessagingProtocol) handleIncomingStream(stream libp2pnetwork.Stream) {
	defer stream.Close()

	reader := bufio.NewReader(stream)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err != io.EOF {
				log.Printf("Error reading from stream: %v", err)
			}
			return
		}

		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		var msg Message
		if err := json.Unmarshal(line, &msg); err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			continue
		}

		// Check if this is a response to a sent request.
		if ch, ok := mp.sentRequests.Load(msg.TransactionID); ok {
			if responseCh, ok := ch.(chan *Message); ok {
				responseCh <- &msg
				continue
			}
		}

		// Process unknown messages.
		if mp.processor != nil {
			mp.processor(&msg, mp)
		}
	}
}
