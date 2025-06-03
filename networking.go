package distributed_sort

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	mplex "github.com/bino7/go-libp2p/p2p/muxer/mplex"
	hostlib "github.com/libp2p/go-libp2p"
	crypto "github.com/libp2p/go-libp2p/core/crypto"
	noise "github.com/libp2p/go-libp2p/p2p/security/noise"
	ma "github.com/multiformats/go-multiaddr"

	// peerstore "github.com/libp2p/go-libp2p/core/peer"
	host "github.com/libp2p/go-libp2p/core/host"
	network "github.com/libp2p/go-libp2p/core/network"
	protocol "github.com/libp2p/go-libp2p/core/protocol"
)

const (
	ProtocolID = "/p2p/messaging/1.0.0"
)

type Libp2pHost struct {
	// host         hostlib.Host
	host         host.Host
	listenAddr   ma.Multiaddr
	msgProcessor func(map[string]interface{})
}

func NewLibp2pHost(port int, msgProcessor func(map[string]interface{})) (*Libp2pHost, error) {
	_ = context.Background()
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
	l.SendMessage(message)
	msgMap := message.(map[string]interface{}) // unsafe cast; better with a defined interface
	l.msgProcessor(msgMap)
}

func (l *Libp2pHost) SendMessage(message interface{}) {
	encoded, err := json.Marshal(message)
	if err != nil {
		log.Println("Failed to encode message:", err)
		return
	}
	fmt.Println("Sending message:", string(encoded))
	// actual sending to peers not implemented here (requires peer ID, streams, etc.)
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
