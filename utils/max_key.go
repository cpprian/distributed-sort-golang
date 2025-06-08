package utils

import "github.com/cpprian/distributed-sort-golang/neighbours"

func MaxKey(m map[int64]neighbours.Neighbour) int64 {
	maxID := int64(0)
	for id := range m {
		if id > maxID {
			maxID = id
		}
	}
	return maxID
}