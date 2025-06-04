package networking

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	hostlib "github.com/libp2p/go-libp2p"
	mplex "github.com/libp2p/go-libp2p-mplex"
	crypto "github.com/libp2p/go-libp2p/core/crypto"
	noise "github.com/libp2p/go-libp2p/p2p/security/noise"
	ma "github.com/multiformats/go-multiaddr"

	host "github.com/libp2p/go-libp2p/core/host"
	network "github.com/libp2p/go-libp2p/core/network"
	peer "github.com/libp2p/go-libp2p/core/peer"
	protocol "github.com/libp2p/go-libp2p/core/protocol"

	"github.com/cpprian/distributed-sort-golang/messages"
)

const (
	ProtocolID = "/p2p/messaging/1.0.0"
)

type Libp2pHost struct {
	host         host.Host
	listenAddr   ma.Multiaddr
	msgProcessor func(map[string]interface{})
	outgoing     chan string
}

func NewLibp2pHost(port int64, msgProcessor func(map[string]interface{})) (*Libp2pHost, error) {
	listenAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port))

	prvKey, _, err := crypto.GenerateEd25519Key(nil)
	if err != nil {
		return nil, err
	}

	h, err := hostlib.New(
		hostlib.Identity(prvKey),
		hostlib.ListenAddrs(listenAddr),
		hostlib.Security(noise.ID, noise.New),
		hostlib.Muxer(mplex.ID, mplex.DefaultTransport),
	)
	if err != nil {
		return nil, err
	}

	host := &Libp2pHost{
		host:         h,
		listenAddr:   listenAddr,
		msgProcessor: msgProcessor,
		outgoing:     make(chan string, 100),
	}

	h.SetStreamHandler(protocol.ID(ProtocolID), host.streamHandler)

	addrInfo := peer.AddrInfo{
		ID:    h.ID(),
		Addrs: h.Addrs(),
	}
	fullAddrs, err := peer.AddrInfoToP2pAddrs(&addrInfo)
	if err != nil || len(fullAddrs) == 0 {
		return nil, fmt.Errorf("failed to build full multiaddr: %w", err)
	}

	fmt.Println("Your full multiaddr is:", fullAddrs[0].String())

	return host, nil
}

func (l *Libp2pHost) streamHandler(s network.Stream) {
	log.Println("New stream opened")
	go l.readData(s)
	go l.writeData(s)
}

func (l *Libp2pHost) readData(s network.Stream) {
	r := bufio.NewReader(s)
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			log.Println("Error reading from stream:", err)
			s.Close()
			return
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		log.Println("Received message:", line)

		var msg map[string]interface{}
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			log.Println("Failed to decode message:", err)
			continue
		}

		if l.msgProcessor != nil {
			l.msgProcessor(msg)
		}
	}
}

func (l *Libp2pHost) writeData(s network.Stream) {
	w := bufio.NewWriter(s)
	for msg := range l.outgoing {
		w.WriteString(msg + "\n")
		w.Flush()
	}
}

func (l *Libp2pHost) Broadcast(message interface{}) {
	encoded, err := json.Marshal(message)
	if err != nil {
		log.Println("Failed to encode message:", err)
		return
	}
	fmt.Println("Broadcasting message:", string(encoded))
	log.Println("Broadcast to peers: ", l.host.Network().Peers())
	for _, peer := range l.host.Network().Peers() {
		log.Println("Sending message to peer:", peer)	
		// Create a new stream to the peer
		stream, err := l.host.NewStream(context.Background(), peer, protocol.ID(ProtocolID))
		if err != nil {
			log.Println("Failed to create stream:", err)
			continue
		}
		// defer stream.Close()
		w := bufio.NewWriter(stream)
		w.WriteString(string(encoded) + "\n")
		w.Flush()
	}
}

func (l *Libp2pHost) Connect(ctx context.Context, targetAddr ma.Multiaddr) error {
	peerInfo, err := peer.AddrInfoFromP2pAddr(targetAddr)
	if err != nil {
		return fmt.Errorf("failed to parse multiaddr: %w", err)
	}

	if err := l.host.Connect(ctx, *peerInfo); err != nil {
		return fmt.Errorf("failed to connect to peer: %w", err)
	}

	log.Println("Connected to peer:", peerInfo.ID)
	return nil
}

func (l *Libp2pHost) SendMessage(msg messages.MessageInterface) {
	encoded, err := json.Marshal(msg)
	if err != nil {
		log.Println("Failed to encode message:", err)
		return
	}
	fmt.Println("Sending message:", string(encoded))

	for _, peer := range l.host.Network().Peers() {
		stream, err := l.host.NewStream(context.Background(), peer, protocol.ID(ProtocolID))
		if err != nil {
			log.Println("Failed to create stream:", err)
			continue
		}
		w := bufio.NewWriter(stream)
		w.WriteString(string(encoded) + "\n")
		w.Flush()
		stream.Close()
	}
	log.Println("Message sent to all peers")
}

func (l *Libp2pHost) GetListenAddress() string {
	return l.listenAddr.String()
}

func (l *Libp2pHost) RunForever() {
	log.Println("Host running at:", l.GetListenAddress())
	select {}
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

