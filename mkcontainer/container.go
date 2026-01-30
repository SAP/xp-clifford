package mkcontainer

import (
	"iter"
	"maps"
	"slices"
	"sync"
)

type Container interface {
	Store(obj ...Object)

	GetByGUID(guid string) ObjectWithGUID
	GetGUIDs() []string
	AllByGUIDs() iter.Seq2[string, ObjectWithGUID]

	GetByName(name string) []ObjectWithName
	GetNames() []string
	AllByNames() iter.Seq2[string, []ObjectWithName]
}

type container struct {
	lock      sync.RWMutex
	guidIndex map[string]ObjectWithGUID
	nameIndex map[string][]ObjectWithName
}

func New() Container {
	return &container{
		guidIndex: make(map[string]ObjectWithGUID),
		nameIndex: make(map[string][]ObjectWithName),
	}
}

var _ Container = &container{}

func (c *container) Store(objects ...Object) {
	c.lock.Lock()
	defer c.lock.Unlock()
	for _, obj := range objects {
		if owg, ok := obj.(ObjectWithGUID); ok {
			c.guidIndex[owg.GetGUID()] = owg
		}
		if own, ok := obj.(ObjectWithName); ok {
			name := own.GetName()
			c.nameIndex[name] = append(c.nameIndex[name], own)
		}
	}
}

func (c *container) GetByGUID(guid string) ObjectWithGUID {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.guidIndex[guid]
}

func (c *container) GetGUIDs() []string {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return slices.Sorted(maps.Keys(c.guidIndex))
}

func (c *container) AllByGUIDs() iter.Seq2[string, ObjectWithGUID] {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return maps.All(c.guidIndex)
}

func (c *container) GetByName(name string) []ObjectWithName {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.nameIndex[name]
}

func (c *container) GetNames() []string {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return slices.Sorted(maps.Keys(c.nameIndex))
}

func (c *container) AllByNames() iter.Seq2[string, []ObjectWithName] {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return maps.All(c.nameIndex)
}
