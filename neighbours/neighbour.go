package neighbours

import (
	ma "github.com/multiformats/go-multiaddr"
)

type Neighbour struct {
	Multiaddr ma.Multiaddr `json:"multiaddr"`
	ID        int64        `json:"id"`
}

func NewNeighbour(multiaddr ma.Multiaddr, id int64) *Neighbour {
	return &Neighbour{
		Multiaddr: multiaddr,
		ID:        id,
	}
}

func (n *Neighbour) GetMultiaddr() ma.Multiaddr {
	return n.Multiaddr
}

func (n *Neighbour) GetID() int64 {
	return n.ID
}
