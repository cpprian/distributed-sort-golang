package neighbour

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

func (n Neighbour) String() string {
	return fmt.Sprintf("Neighbour(ID: %d, Multiaddr: %s)", n.ID, n.Multiaddr.String())
}

func (n Neighbour) Equal(other Neighbour) bool {
	return n.ID == other.ID && n.Multiaddr.Equal(other.Multiaddr)
}

func (n Neighbour) IsValid() bool {
	return n.ID > 0 && n.Multiaddr != nil
}

func (n Neighbour) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"multiaddr": n.Multiaddr.String(),
		"id":        n.ID,
	}
}

func NeighbourFromMap(data map[string]interface{}) (Neighbour, error) {
	addrStr, ok := data["multiaddr"].(string)
	if !ok {
		return Neighbour{}, fmt.Errorf("invalid multiaddr type")
	}

	id, ok := data["id"].(int64)
	if !ok {
		return Neighbour{}, fmt.Errorf("invalid id type")
	}

	addr, err := multiaddr.NewMultiaddr(addrStr)
	if err != nil {
		return Neighbour{}, fmt.Errorf("invalid multiaddr: %w", err)
	}

	return Neighbour{
		Multiaddr: addr,
		ID:        id,
	}, nil
}
