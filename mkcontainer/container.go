package mkcontainer

import (
	"iter"
	"maps"
	"slices"
	"sync"
)

// Container provides thread-safe storage and retrieval of items with
// automatic indexing by GUID and name. Items are indexed based on which
// interfaces they implement (ItemWithGUID and/or ItemWithName).
type Container interface {
	// Store adds one or more items to the container. Each item is
	// automatically indexed by its GUID and/or name if it implements
	// the corresponding interface.
	Store(obj ...Item)

	// GetByGUID retrieves an item by its unique GUID.
	// Returns nil if no item with the given GUID exists.
	GetByGUID(guid string) ItemWithGUID

	// GetGUIDs returns a sorted slice of all GUIDs in the container.
	GetGUIDs() []string

	// AllByGUIDs returns an iterator over all GUID-indexed items
	// as key-value pairs (GUID -> item).
	AllByGUIDs() iter.Seq2[string, ItemWithGUID]

	// GetByName retrieves all items sharing the given name.
	// Returns nil if no items with the given name exist.
	GetByName(name string) []ItemWithName

	// GetNames returns a sorted slice of all unique names in the container.
	GetNames() []string

	// AllByNames returns an iterator over all name-indexed items
	// as key-value pairs (name -> slice of items).
	AllByNames() iter.Seq2[string, []ItemWithName]

	// IsEmpty returns true if the container has no items stored,
	// false otherwise.
	IsEmpty() bool
}

// container is the concrete implementation of Container.
// It uses separate maps for GUID and name indexing, protected by a
// read-write mutex to ensure thread safety.
type container struct {
	lock      sync.RWMutex
	guidIndex map[string]ItemWithGUID   // maps GUID -> single item
	nameIndex map[string][]ItemWithName // maps name -> multiple items
}

// New creates and returns an empty Container ready for use.
func New() Container {
	return &container{
		guidIndex: make(map[string]ItemWithGUID),
		nameIndex: make(map[string][]ItemWithName),
	}
}

// Compile-time check that container implements Container.
var _ Container = &container{}

func (c *container) Store(objects ...Item) {
	c.lock.Lock()
	defer c.lock.Unlock()
	for _, obj := range objects {
		if owg, ok := obj.(ItemWithGUID); ok {
			c.guidIndex[owg.GetGUID()] = owg
		}
		if own, ok := obj.(ItemWithName); ok {
			name := own.GetName()
			c.nameIndex[name] = append(c.nameIndex[name], own)
		}
	}
}

func (c *container) GetByGUID(guid string) ItemWithGUID {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.guidIndex[guid]
}

func (c *container) GetGUIDs() []string {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return slices.Sorted(maps.Keys(c.guidIndex))
}

func (c *container) AllByGUIDs() iter.Seq2[string, ItemWithGUID] {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return maps.All(c.guidIndex)
}

func (c *container) GetByName(name string) []ItemWithName {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.nameIndex[name]
}

func (c *container) GetNames() []string {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return slices.Sorted(maps.Keys(c.nameIndex))
}

func (c *container) AllByNames() iter.Seq2[string, []ItemWithName] {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return maps.All(c.nameIndex)
}

// IsEmpty returns true if the container has no items stored,
// false otherwise.
func (c *container) IsEmpty() bool {
	c.lock.RLock()
	defer c.lock.RUnlock()

	return len(c.nameIndex) == 0 && len(c.guidIndex) == 0
}
