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
	"github.com/cpprian/distributed-sort-golang/networks"
	"github.com/cpprian/distributed-sort-golang/sorting"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background()) // Create a cancelable context

	sortingManager := sorting.NewSortingManager()

	messageHandler := func(data map[string]interface{}) {
		log.Println("Received message:", data)
		rawType, ok := data["messageType"].(string)
		if !ok {
			log.Println("Invalid messageType (not string):", data["messageType"])
			return
		}

		msgType := messages.MessageType(rawType)
		log.Println("Message type:", msgType)

		var msg messages.MessageInterface
		msg, err := messages.DeserializeMessage(data, msgType)
		if err != nil {
			log.Println("Failed to deserialize message:", err)
			return
		}
		log.Println("Deserialized message:", msg)
		sortingManager.ProcessMessage(msg)
	}

	h, err := networks.NewLibp2pHost(0, messageHandler)
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

	h.Start(ctx)

	fmt.Println("Activated node:", h.Addrs(), h.ID())

	for range 3 {
		sortingManager.Add(int64(rand.Intn(100)))
	}

	log.Println("Initial items:", sortingManager.GetItems())

	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := scanner.Text()
			log.Println("Received command:", line)
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
				cancel()
				break
			} else {
				fmt.Println("Unknown command:", line)
			}
		}
	}()

	// Wait for context to be done
	<-ctx.Done()
	log.Println("Context done, stopping host...")
	h.Stop()
	log.Println("Shutting down...")
}
