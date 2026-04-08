# mkcontainer — Multi-Key Container

The `mkcontainer` package provides a thread-safe container for storing and retrieving items indexed by multiple keys.

## Features

- **Multi-key indexing** — Items indexed by GUID, name, or both
- **Thread-safe** — All operations safe for concurrent use
- **O(1) lookups** — By GUID (unique) and by name (non-unique)

## Interfaces

Items implement one or both interfaces:

| Interface      | Method             | Uniqueness     | Lookup Returns |
|----------------|--------------------|----------------|----------------|
| `ItemWithGUID` | `GetGUID() string` | Must be unique | Single item    |
| `ItemWithName` | `GetName() string` | Not unique     | Slice of items |

## Usage

```go
type Document struct {
    ID    string
    Title string
}

func (d *Document) GetGUID() string { return d.ID }
func (d *Document) GetName() string { return d.Title }

c := mkcontainer.New()

c.Store(
    &Document{ID: "doc-1", Title: "Report"},
    &Document{ID: "doc-2", Title: "Report"},
    &Document{ID: "doc-3", Title: "Summary"},
)

// Lookup by unique GUID
doc := c.GetByGUID("doc-1")

// Lookup all items with the same name
reports := c.GetByName("Report")

// Iterate all GUIDs (sorted)
for _, guid := range c.GetGUIDs() {
    fmt.Printf("GUID: %s\n", guid)
}

// Iterate all items by GUID
for guid, item := range c.AllByGUIDs() {
    fmt.Printf("%s -> %s\n", guid, item.(*Document).Title)
}
```
