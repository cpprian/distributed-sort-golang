package networking

import (
	"context"
	"fmt"
	"log"

	host "github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
)

func (l *Libp2pHost) GetListenAddress() string {
	return l.listenAddr.String()
}

func (l *Libp2pHost) Start(ctx context.Context) error {
	if err := l.host.Network().Listen(l.listenAddr); err != nil {
		return fmt.Errorf("failed to start host: %w", err)
	}
	log.Println("Host started successfully")
	return nil
}

func (l *Libp2pHost) Stop() error {
	if err := l.host.Close(); err != nil {
		return fmt.Errorf("failed to stop host: %w", err)
	}
	log.Println("Host stopped successfully")
	return nil
}

func (l *Libp2pHost) Close() error {
	if err := l.host.Close(); err != nil {
		return fmt.Errorf("failed to close host: %w", err)
	}
	log.Println("Host closed successfully")
	return nil
}

func (l *Libp2pHost) GetHost() host.Host {
	return l.host
}

func (l *Libp2pHost) GetPeerID() peer.ID {
	return l.host.ID()
}

func (l *Libp2pHost) GetAddrs() []ma.Multiaddr {
	addrs := make([]ma.Multiaddr, 0, len(l.host.Addrs()))
	addrs = append(addrs, l.host.Addrs()...)
	return addrs
}

func (l *Libp2pHost) GetPeerInfo() map[string][]string {
	peerInfo := make(map[string][]string)
	for _, peer := range l.host.Network().Peers() {
		addresses := make([]string, 0, len(l.host.Peerstore().Addrs(peer)))
		for _, addr := range l.host.Peerstore().Addrs(peer) {
			addresses = append(addresses, addr.String())
		}
		peerInfo[peer.String()] = addresses
	}
	return peerInfo
}

func (l *Libp2pHost) Addrs() []ma.Multiaddr {
	addrs := make([]ma.Multiaddr, 0, len(l.host.Addrs()))
	addrs = append(addrs, l.host.Addrs()...)
	return addrs
}

func (l *Libp2pHost) ID() string {
	return l.host.ID().String()
}
