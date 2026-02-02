// Package mkcontainer provides a thread-safe multi-key container for storing
// and retrieving items indexed by multiple keys simultaneously.
//
// The container automatically indexes items based on which interfaces they
// implement, enabling efficient lookups by different keys:
//
//   - [ItemWithGUID]: Items are indexed by a globally unique identifier,
//     providing O(1) lookups. Each GUID must be unique within the container.
//
//   - [ItemWithName]: Items are indexed by name, with multiple items allowed
//     to share the same name. Lookups return all items matching the name.
//
// An item may implement both interfaces to be indexed by both GUID and name.
//
// # Thread Safety
//
// All container operations are safe for concurrent use. The implementation
// uses a read-write mutex, allowing multiple concurrent readers while
// serializing write operations.
//
// # Example Usage
//
//	type Document struct {
//	    ID    string
//	    Title string
//	}
//
//	func (d *Document) GetGUID() string { return d.ID }
//	func (d *Document) GetName() string { return d.Title }
//
//	func main() {
//	    c := mkcontainer.New()
//
//	    c.Store(
//	        &Document{ID: "doc-1", Title: "Report"},
//	        &Document{ID: "doc-2", Title: "Report"},
//	        &Document{ID: "doc-3", Title: "Summary"},
//	    )
//
//	    // Lookup by unique GUID
//	    doc := c.GetByGUID("doc-1")
//
//	    // Lookup all documents named "Report"
//	    reports := c.GetByName("Report") // returns 2 items
//	}
package mkcontainer
