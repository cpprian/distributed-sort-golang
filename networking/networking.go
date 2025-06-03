package networking

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	hostlib "github.com/libp2p/go-libp2p"
	mplex "github.com/libp2p/go-libp2p-mplex"
	crypto "github.com/libp2p/go-libp2p/core/crypto"
	noise "github.com/libp2p/go-libp2p/p2p/security/noise"
	ma "github.com/multiformats/go-multiaddr"

	host "github.com/libp2p/go-libp2p/core/host"
	network "github.com/libp2p/go-libp2p/core/network"
	protocol "github.com/libp2p/go-libp2p/core/protocol"
)

const (
	ProtocolID = "/p2p/messaging/1.0.0"
)

type Libp2pHost struct {
	host         host.Host
	listenAddr   ma.Multiaddr
	msgProcessor func(map[string]interface{})
}

func NewLibp2pHost(port int, msgProcessor func(map[string]interface{})) (*Libp2pHost, error) {
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
	}

	h.SetStreamHandler(protocol.ID(ProtocolID), host.streamHandler)
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
			log.Println("Read error:", err)
			return
		}
		line = strings.TrimSpace(line)
		if line != "" {
			var msg map[string]interface{}
			err := json.Unmarshal([]byte(line), &msg)
			if err != nil {
				log.Println("JSON decode error:", err)
				continue
			}
			l.msgProcessor(msg)
		}
	}
}

func (l *Libp2pHost) writeData(s network.Stream) {
	scanner := bufio.NewScanner(os.Stdin)
	w := bufio.NewWriter(s)
	for scanner.Scan() {
		text := scanner.Text()
		if text != "" {
			w.WriteString(text + "\n")
			w.Flush()
		}
	}
}

func (l *Libp2pHost) Broadcast(message interface{}) {
	encoded, err := json.Marshal(message)
	if err != nil {
		log.Println("Failed to encode message:", err)
		return
	}
	fmt.Println("Broadcasting message:", string(encoded))
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
}

func (l *Libp2pHost) SendMessage(message interface{}) {
	encoded, err := json.Marshal(message)
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

func (l *Libp2pHost) GetPeerID() string {
	return l.host.ID().String()
}

func (l *Libp2pHost) GetPeerAddresses() []string {
	addresses := make([]string, 0, len(l.host.Addrs()))
	for _, addr := range l.host.Addrs() {
		addresses = append(addresses, addr.String())
	}
	return addresses
}

func (l *Libp2pHost) GetPeerIDAndAddresses() map[string][]string {
	peerInfo := make(map[string][]string)
	peerID := l.GetPeerID()
	peerInfo[peerID] = l.GetPeerAddresses()
	return peerInfo
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
