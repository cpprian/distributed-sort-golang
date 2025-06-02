package serializers

import (
	"encoding/json"
	"fmt"

	ma "github.com/multiformats/go-multiaddr"
)

// MultiaddrJSON to alias typu Multiaddr z implementacją niestandardowego JSON marshallera i unmarshallera
type MultiaddrJSON struct {
	ma.Multiaddr
}

// UnmarshalJSON zamienia ciąg znaków JSON na Multiaddr
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

// MarshalJSON konwertuje Multiaddr do postaci JSON (jako string)
func (m MultiaddrJSON) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.Multiaddr.String())
}
