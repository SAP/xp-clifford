package mkcontainer

// Item represents any type that can be stored in a Container.
// This is an alias for the empty interface, allowing maximum flexibility
// in what can be stored.
type Item any

// ItemWithGUID defines an interface for items that have a globally unique
// identifier. Items implementing this interface will be indexed by their
// GUID, enabling O(1) lookups.
type ItemWithGUID interface {
	// GetGUID returns the globally unique identifier for this item.
	// The GUID must be unique across all items in the container.
	GetGUID() string
}

// ItemWithName defines an interface for items that have a name.
// Unlike GUIDs, names are not required to be unique; multiple items
// can share the same name and will be grouped together in the index.
type ItemWithName interface {
	// GetName returns the name of this item.
	// Multiple items may share the same name.
	GetName() string
}
