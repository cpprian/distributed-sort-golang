package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"

	"github.com/cpprian/distributed-sort-golang/networks"
	"github.com/cpprian/distributed-sort-golang/sorting"

	ma "github.com/multiformats/go-multiaddr"
)

func main() {
	log.Println("Starting distributed sorting application...")

	sortingManager := sorting.NewSortingManager()
	messaging := networks.NewMessagingProtocol(sortingManager.ProcessMessage)
	sortingManager.SetMessagingController(messaging)

	host, err := networks.NewLibp2pHost(*messaging)
	if err != nil {
		log.Fatalf("Failed to create libp2p host: %v", err)
		return
	}

	sortingManager.SetHost(*host)
	log.Println("Host set with address:", host.Host.Network().ListenAddresses()[0])

	var targetAddr ma.Multiaddr
	if len(os.Args) > 1 {
		var err error
		targetAddr, err = ma.NewMultiaddr(os.Args[1])
		if err != nil {
			log.Printf("Invalid multiaddress: %s, error: %v\n", os.Args[1], err)
			return
		}
		log.Println("Target address provided:", targetAddr)
	} else {
		log.Println("No target address provided. Activating with self ID and address.")
	}

	sortingManager.Activate(targetAddr)
	log.Println("Activated node with ID:", sortingManager.ID, "and address:", sortingManager.Host.Host.Network().ListenAddresses()[0])

	for range 10 {
		sortingManager.AddItem(int64(rand.Intn(100)))
	}

	log.Println("Initial items:", sortingManager.GetItems())

	go func() {
		for {
			var input string
			_, err := fmt.Scanln(&input)
			if err != nil {
				log.Println("Error reading input:", err)
				continue
			}
			log.Println("Received command:", input)

			switch strings.ToLower(input) {
			case "allitems":
				fmt.Println(sortingManager.GetAllItems())
			case "items":
				fmt.Println(sortingManager.GetItems())
			case "nodes":
				for id, addr := range sortingManager.ParticipatingNodes {
					fmt.Printf("Node ID: %d, Address: %v\n", id, addr)
				}
			default:
				if strings.HasPrefix(input, "add ") {
					numStr := strings.TrimSpace(strings.TrimPrefix(input, "add "))
					num, err := strconv.ParseInt(numStr, 10, 64)
					if err == nil {
						sortingManager.AddItem(num)
					} else {
						log.Println("Invalid number format for add command:", numStr)
					}
				} else if strings.HasPrefix(input, "del ") {
					numStr := strings.TrimSpace(strings.TrimPrefix(input, "del "))
					num, err := strconv.ParseInt(numStr, 10, 64)
					if err == nil {
						sortingManager.RemoveItem(num)
					} else {
						log.Println("Invalid number format for del command:", numStr)
					}
				} else {
					log.Println("Unknown command:", input)
				}
			}
		}
	}()
	log.Println("Distributed sorting application is running. Type 'allitems', 'items', 'add <number>', 'del <number>', or 'nodes' to interact.")
	select {} // Keep the main goroutine running
}
