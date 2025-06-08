package networks

import (
	"context"
	"log"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"

	ma "github.com/multiformats/go-multiaddr"
)

func DialMessagingController(h host.Host, peerID peer.ID, protocolID protocol.ID, processor UnknownMessageProcessor) (MessagingController, error) {
	log.Println("Dialing peer:", peerID, "with protocol:", protocolID)

	stream, err := h.NewStream(context.Background(), peerID, protocolID)
	if err != nil {
		log.Println("Error dialing peer:", err)
		return nil, err
	}

	controller := NewMessagingInitiator(processor, stream)
	log.Println("Successfully dialed peer:", peerID)
	return controller, nil
}

func DialByMultiaddr(h host.Host, addr ma.Multiaddr, protocolID protocol.ID, processor UnknownMessageProcessor) (MessagingController, error) {
	log.Println("Dialing multiaddr:", addr, "with protocol:", protocolID)

	info, err := peer.AddrInfoFromP2pAddr(addr)
	if err != nil {
		log.Println("Error getting peer info from multiaddr:", err)
		return nil, err
	}
	if err := h.Connect(context.Background(), *info); err != nil {
		log.Println("Error connecting to peer:", err)
		return nil, err
	}
	return DialMessagingController(h, info.ID, protocolID, processor)
}
