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
	ms "github.com/cpprian/distributed-sort-golang/messages"
	serial "github.com/cpprian/distributed-sort-golang/serializers"
)

func main() {
	ctx := context.Background()

	// Create SortingManager instance
	sortingManager := distributed_sort.NewSortingManager()

	// Create a Libp2pHost instance
	host, err := distributed_sort.NewLibp2pHost(0, func(msg map[string]interface{}) {
		var messageType string
		if val, ok := msg["type"]; ok {
			messageType = val.(string)
		}

		switch messageType {
		case ms.AnnounceSelfMessageType:
			var announceMsg ms.AnnounceSelfMessage
			if err := serial.Unmarshal(msg, &announceMsg); err != nil {
				log.Printf("Error unmarshalling AnnounceSelfMessage: %v", err)
				return
			}
			sortingManager.ParticipatingNodes[announceMsg.ID] = announceMsg.ListeningAddress

		case ms.AddItemMessageType:
			var addItemMsg ms.AddItemMessage
			if err := serial.Unmarshal(msg, &addItemMsg); err != nil {
				log.Printf("Error unmarshalling AddItemMessage: %v", err)
				return
			}
			sortingManager.Add(addItemMsg.Item)

		case ms.RemoveItemMessageType:
			var removeItemMsg ms.RemoveItemMessage
			if err := serial.Unmarshal(msg, &removeItemMsg); err != nil {
				log.Printf("Error unmarshalling RemoveItemMessage: %v", err)
				return
			}
			sortingManager.Remove(removeItemMsg.Item)

		default:
			log.Printf("Unknown message type: %s", messageType)
		}
	})
	if err != nil {
		log.Fatalf("Failed to create Libp2p host: %v", err)
	}

	sortingManager.SetHost(host)
	// Start the host
	if err := host.Start(ctx); err != nil {
		log.Fatalf("Failed to start host: %v", err)
	}

	// Print the listening address
	listenAddr := host.GetListenAddress()
	log.Printf("Listening on: %s", listenAddr)
	// Announce self to the network
	sortingManager.AnnounceSelf()
	// Start reading input from the user
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter a number to add (or 'exit' to quit): ")
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Error reading input: %v", err)
			continue
		}
		input = strings.TrimSpace(input)

		if input == "exit" {
			break
		}

		item, err := strconv.Atoi(input)
		if err != nil {
			log.Printf("Invalid input, please enter a valid number: %v", err)
			continue
		}

		sortingManager.Add(item)
		log.Printf("Current sorted items: %v", sortingManager.Items)
	}
	log.Println("Exiting...")
	if err := host.Stop(); err != nil {
		log.Fatalf("Failed to stop host: %v", err)
	}
	if err := host.Close(); err != nil {
		log.Fatalf("Failed to close host: %v", err)
	}
	log.Println("Host closed successfully.")
	fmt.Println("Goodbye!")
}
