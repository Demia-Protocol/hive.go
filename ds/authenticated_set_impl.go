package ds

import (
	"github.com/iotaledger/hive.go/ds/types"
	"github.com/iotaledger/hive.go/kvstore"
)

// Set is a sparse merkle tree based set.
type authenticatedSet[K any] struct {
	*authenticatedMap[K, types.Empty]
}

// NewSet creates a new sparse merkle tree based set.
func newAuthenticatedSet[K any](store kvstore.KVStore, kToBytes kvstore.ObjectToBytes[K], bytesToK kvstore.BytesToObject[K]) *authenticatedSet[K] {
	return &authenticatedSet[K]{
		authenticatedMap: newAuthenticatedMap(store, kToBytes, bytesToK, types.Empty.Bytes, types.EmptyFromBytes),
	}
}

// Add adds the key to the set.
func (s *authenticatedSet[K]) Add(key K) error {
	return s.Set(key, types.Void)
}

// Stream iterates over the set and calls the callback for each element.
func (s *authenticatedSet[K]) Stream(callback func(key K) error) error {
	return s.authenticatedMap.Stream(func(key K, _ types.Empty) error {
		return callback(key)
	})
}
