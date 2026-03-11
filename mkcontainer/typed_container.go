package mkcontainer

import (
	"iter"
	"maps"
	"slices"
	"sync"
)

// TypedContainer provides thread-safe storage and retrieval of items
// with automatic indexing by GUID and name. Items are indexed based
// on which interfaces they implement (ItemWithGUID and/or
// ItemWithName).
type TypedContainer[T any] interface {
	// Store adds one or more items to the container. Each item is
	// automatically indexed by its GUID and/or name if it implements
	// the corresponding interface.
	Store(obj ...T)

	// GetByGUID retrieves an item by its unique GUID.
	// Returns nil if no item with the given GUID exists.
	GetByGUID(guid string) T

	// GetGUIDs returns a sorted slice of all GUIDs in the container.
	GetGUIDs() []string

	// AllByGUIDs returns an iterator over all GUID-indexed items
	// as key-value pairs (GUID -> item).
	AllByGUIDs() iter.Seq2[string, T]

	// GetByName retrieves all items sharing the given name.
	// Returns nil if no items with the given name exist.
	GetByName(name string) []T

	// GetNames returns a sorted slice of all unique names in the container.
	GetNames() []string

	// AllByNames returns an iterator over all name-indexed items
	// as key-value pairs (name -> slice of items).
	AllByNames() iter.Seq2[string, []T]

	// IsEmpty returns true if the container has no items stored,
	// false otherwise.
	IsEmpty() bool
}

// typedContainer is the concrete implementation of Container. It
// uses separate maps for GUID and name indexing, protected by a
// read-write mutex to ensure thread safety.
type typedContainer[T any] struct {
	lock      sync.RWMutex
	guidIndex map[string]T   // maps GUID -> single item
	nameIndex map[string][]T // maps name -> multiple items
}

// NewTyped creates and returns an empty TypedContainer ready for use.
func NewTyped[T any]() TypedContainer[T] {
	return &typedContainer[T]{
		guidIndex: make(map[string]T),
		nameIndex: make(map[string][]T),
	}
}

// Compile-time check that typedContainer implements TypedContainer.
var _ TypedContainer[int] = &typedContainer[int]{}

func (c *typedContainer[T]) Store(objects ...T) {
	c.lock.Lock()
	defer c.lock.Unlock()
	for _, obj := range objects {
		if owg, ok := any(obj).(ItemWithGUID); ok {
			c.guidIndex[owg.GetGUID()] = owg.(T)
		}
		if own, ok := any(obj).(ItemWithName); ok {
			name := own.GetName()
			c.nameIndex[name] = append(c.nameIndex[name], own.(T))
		}
	}
}

func (c *typedContainer[T]) GetByGUID(guid string) T {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.guidIndex[guid]
}

func (c *typedContainer[T]) GetGUIDs() []string {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return slices.Sorted(maps.Keys(c.guidIndex))
}

func (c *typedContainer[T]) AllByGUIDs() iter.Seq2[string, T] {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return maps.All(c.guidIndex)
}

func (c *typedContainer[T]) GetByName(name string) []T {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.nameIndex[name]
}

func (c *typedContainer[T]) GetNames() []string {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return slices.Sorted(maps.Keys(c.nameIndex))
}

func (c *typedContainer[T]) AllByNames() iter.Seq2[string, []T] {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return maps.All(c.nameIndex)
}

func (c *typedContainer[T]) IsEmpty() bool {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return len(c.guidIndex) == 0 && len(c.nameIndex) == 0
}
