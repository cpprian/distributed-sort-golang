package neighbours

import (
	ma "github.com/multiformats/go-multiaddr"
)

type Neighbour struct {
	Multiaddr ma.Multiaddr `json:"multiaddr"`
	ID        int64        `json:"id"`
}
