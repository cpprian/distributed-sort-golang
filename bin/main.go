package main

import (
	"bufio"
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

	host, err := networks.NewLibp2pHost(messaging)
	if err != nil {
		log.Fatalf("Failed to create libp2p host: %v", err)
		return
	}

	sortingManager.SetHost(host)
	log.Println("Host set with address:", host.Host.Addrs(), "ID:", host.Host.ID())

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
	log.Println("Activated node with ID:", sortingManager.ID, "and address:", sortingManager.Host.Host.Addrs())
	
	for range 3 {
		sortingManager.AddItem(int64(rand.Intn(100)))
	}

	log.Println("Initial items:", sortingManager.GetItems())

	scanner := bufio.NewScanner(os.Stdin)
	go func() {
		for {
			var input string
			fmt.Print("Enter command (allitems, items, add <number>, del <number>, nodes, id): ")
			if !scanner.Scan() {
				log.Println("Error reading input:", scanner.Err())
				continue
			}
			input = scanner.Text()
			if input == "" {
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
			case "id":
				fmt.Println("Node ID:", sortingManager.ID)
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
	select {} // Keep the main goroutine running
}
