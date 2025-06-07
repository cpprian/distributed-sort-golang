package sorting

import (
	"sync"

	host "github.com/libp2p/go-libp2p/core/host"
)

type SortingManager struct {
	ID                 int64
	Items              []int64
	ParticipatingNodes map[int64]string
	Host               host.Host
	mu                 sync.Mutex
}
