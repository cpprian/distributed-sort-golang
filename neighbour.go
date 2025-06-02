package distributed_sort

import (
	"encoding/json"
	"fmt"
	"github.com/multiformats/go-multiaddr"
)

type Neighbour struct {
	Multiaddr multiaddr.Multiaddr `json:"multiaddr"`
	ID        int64               `json:"id"`
}

func (n Neighbour) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Multiaddr string `json:"multiaddr"`
		ID        int64  `json:"id"`
	}{
		Multiaddr: n.Multiaddr.String(),
		ID:        n.ID,
	})
}

func (n *Neighbour) UnmarshalJSON(data []byte) error {
	aux := struct {
		Multiaddr string `json:"multiaddr"`
		ID        int64  `json:"id"`
	}{}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	addr, err := multiaddr.NewMultiaddr(aux.Multiaddr)
	if err != nil {
		return fmt.Errorf("invalid multiaddr: %w", err)
	}

	n.Multiaddr = addr
	n.ID = aux.ID
	return nil
}
