package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"

	libp2p "github.com/libp2p/go-libp2p"
	// "github.com/libp2p/go-libp2p-core/host"
	// "github.com/libp2p/go-libp2p-core/peer"
	// "github.com/libp2p/go-libp2p-core/peerstore"
	ma "github.com/multiformats/go-multiaddr"
	noise "github.com/libp2p/go-libp2p-noise"
	mplex "github.com/libp2p/go-libp2p-mplex"

	"github.com/cpprian/distributed-sort-golang"
)

func main() {
	ctx := context.Background()

	// Create SortingManager instance
	sortingManager := distributed_sort.NewSortingManager[int]()

	// Create Messaging and bind the processMessage function
	messaging := distributed_sort.NewMessagingProtocol(sortingManager.ProcessMessage)
	sortingManager.SetMessaging(messaging)

	// Set up the libp2p host with noise and mplex
	node, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/127.0.0.1/tcp/0"),
		libp2p.Security(noise.ID, noise.New),
		libp2p.Muxer("/mplex/6.7.0", mplex.DefaultTransport),
	)
	if err != nil {
		log.Fatalf("failed to create libp2p host: %v", err)
	}

	sortingManager.SetHost(node)

	// Connect to a peer if address is provided
	var target ma.Multiaddr
	if len(os.Args) > 1 {
		targetStr := os.Args[1]
		target, err = ma.NewMultiaddr(targetStr)
		if err != nil {
			log.Fatalf("invalid multiaddr: %v", err)
		}
	}

	sortingManager.Activate(ctx, target)

	fmt.Printf("Activated node: %v %v\n", node.Addrs(), sortingManager.GetId())

	// Add 10 random items
	r := rand.New(rand.NewSource(0))
	for i := 0; i < 10; i++ {
		sortingManager.Add(r.Intn(100))
	}

	// Command loop
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.EqualFold(line, "allitems"):
			fmt.Println(sortingManager.GetAllItems())
		case strings.EqualFold(line, "items"):
			fmt.Println(sortingManager.GetItems())
		case strings.HasPrefix(strings.ToLower(line), "add "):
			parts := strings.Split(line, " ")
			if len(parts) == 2 {
				if number, err := strconv.Atoi(parts[1]); err == nil {
					sortingManager.Add(number)
				}
			}
		case strings.HasPrefix(strings.ToLower(line), "del "):
			parts := strings.Split(line, " ")
			if len(parts) == 2 {
				if number, err := strconv.Atoi(parts[1]); err == nil {
					sortingManager.Remove(number)
				}
			}
		}
	}
}
