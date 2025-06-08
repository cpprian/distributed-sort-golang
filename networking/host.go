package networking

import (
	"fmt"
	"log"

	hostlib "github.com/libp2p/go-libp2p"
	mplex "github.com/libp2p/go-libp2p-mplex"
	crypto "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	peer "github.com/libp2p/go-libp2p/core/peer"
	protocol "github.com/libp2p/go-libp2p/core/protocol"
	noise "github.com/libp2p/go-libp2p/p2p/security/noise"
	ma "github.com/multiformats/go-multiaddr"
)

const (
	ProtocolID = "/p2p/messaging/1.0.0"
)

type Libp2pHost struct {
	Host host.Host
}

func NewLibp2pHost(messagingProtocol MessagingProtocol) (*Libp2pHost, error) {
	log.Println("Creating new Libp2pHost")
	listenAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", 0))

	prvKey, _, err := crypto.GenerateEd25519Key(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	h, err := hostlib.New(
		hostlib.Identity(prvKey),
		hostlib.ListenAddrs(listenAddr),
		hostlib.Security(noise.ID, noise.New),
		hostlib.Muxer(mplex.ID, mplex.DefaultTransport),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create host: %w", err)
	}

	host := &Libp2pHost{
		Host: h,
	}

	h.SetStreamHandler(protocol.ID(ProtocolID), messagingProtocol.HandleStream)

	addrInfo := peer.AddrInfo{
		ID:    h.ID(),
		Addrs: h.Addrs(),
	}
	fullAddrs, err := peer.AddrInfoToP2pAddrs(&addrInfo)
	if err != nil || len(fullAddrs) == 0 {
		return nil, fmt.Errorf("failed to build full multiaddr: %w", err)
	}

	fmt.Println("Full multiaddr:", fullAddrs[0].String())
	return host, nil
}
