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

	"github.com/multiformats/go-multiaddr"

	"github.com/cpprian/distributed-sort-golang/messages"
	"github.com/cpprian/distributed-sort-golang/networking"
	"github.com/cpprian/distributed-sort-golang/sorting"
)

func main() {
	ctx := context.Background()

	// Create SortingManager
	sortingManager := sorting.NewSortingManager()

	messageHandler := func(data map[string]interface{}) {
		if msg, ok := data["message"].(messages.MessageInterface); ok {
			sortingManager.ProcessMessage(msg)
		}
	}

	h, err := networking.NewLibp2pHost(0, messageHandler)
	if err != nil {
		fmt.Println("Failed to create libp2p host:", err)
		return
	}

	var connectedPeers []multiaddr.Multiaddr
	if len(os.Args) > 1 {
		for _, addrStr := range os.Args[1:] {
			targetAddr, err := multiaddr.NewMultiaddr(addrStr)
			if err != nil {
				fmt.Printf("Invalid multiaddress: %s, error: %v\n", addrStr, err)
				continue
			}
			log.Println("Connecting to peer:", targetAddr)
			if err := h.Connect(ctx, targetAddr); err != nil {
				fmt.Printf("Failed to connect to peer %s: %v\n", addrStr, err)
				continue
			}
			connectedPeers = append(connectedPeers, targetAddr)
		}
	}
	log.Println("Connected peers:", connectedPeers)

	sortingManager.SetHost(h)
	sortingManager.AnnounceSelf()
	for _, peerAddr := range connectedPeers {
		log.Println("Activating peer:", peerAddr)
		sortingManager.Activate(peerAddr.String())
	}
	if err := h.Start(ctx); err != nil {
		fmt.Println("Failed to start host:", err)
		return
	}

	// Output node details
	fmt.Println("Activated node:", h.Addrs(), h.ID())

	for range 3 {
		sortingManager.Add(int64(rand.Intn(100)))
	}

	log.Println("Initial items:", sortingManager.GetItems())

	// CLI
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.EqualFold(line, "allitems") {
			fmt.Println(sortingManager.GetAllItems())
		} else if strings.EqualFold(line, "items") {
			fmt.Println(sortingManager.GetItems())
		} else if strings.HasPrefix(strings.ToLower(line), "add ") {
			parts := strings.Split(line, " ")
			if len(parts) == 2 {
				if num, err := strconv.Atoi(parts[1]); err == nil {
					sortingManager.Add(int64(num))
				}
			}
		} else if strings.HasPrefix(strings.ToLower(line), "del ") {
			parts := strings.Split(line, " ")
			if len(parts) == 2 {
				if num, err := strconv.Atoi(parts[1]); err == nil {
					sortingManager.Remove(int64(num))
				}
			}
		} else if strings.EqualFold(line, "exit") || strings.EqualFold(line, "quit") {
			fmt.Println("Exiting...")
			break
		} else {
			fmt.Println("Unknown command:", line)
		}
	}
}
