package serializers

import (
	"encoding/json"
	"fmt"

	ma "github.com/multiformats/go-multiaddr"
)

type MultiaddrJSON struct {
	ma.Multiaddr
}

func (m *MultiaddrJSON) UnmarshalJSON(data []byte) error {
	var addrStr string
	if err := json.Unmarshal(data, &addrStr); err != nil {
		return fmt.Errorf("invalid JSON string for Multiaddr: %w", err)
	}
	addr, err := ma.NewMultiaddr(addrStr)
	if err != nil {
		return fmt.Errorf("invalid multiaddr: %w", err)
	}
	m.Multiaddr = addr
	return nil
}

func (m MultiaddrJSON) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.Multiaddr.String())
}