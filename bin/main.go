package main

import (
	"bufio"
	"context"
	"fmt"
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

	// Create Libp2p host
	// Wrapper function to adapt the signature
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

	// Connect to peer if multiaddr provided
	var targetAddr multiaddr.Multiaddr
	if len(os.Args) > 1 {
		targetAddr, err = multiaddr.NewMultiaddr(os.Args[1])
		if err != nil {
			fmt.Println("Invalid multiaddress:", err)
			return
		}
	}

	sortingManager.Activate(targetAddr.String())
	if err := h.Start(ctx); err != nil {
		fmt.Println("Failed to start host:", err)
		return
	}

	// Output node details
	fmt.Println("Activated node:", h.Addrs(), h.ID())

	// Populate with 10 random items
	r := rand.New(rand.NewSource(0))
	for i := 0; i < 10; i++ {
		sortingManager.Add(r.Intn(100))
	}

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
					sortingManager.Add(num)
				}
			}
		} else if strings.HasPrefix(strings.ToLower(line), "del ") {
			parts := strings.Split(line, " ")
			if len(parts) == 2 {
				if num, err := strconv.Atoi(parts[1]); err == nil {
					sortingManager.Remove(num)
				}
			}
		}
	}
}
