package mkcontainer_test

import (
	"fmt"

	"github.com/SAP/xp-clifford/mkcontainer"
)

type Document struct {
	ID    string
	Title string
}

func (d *Document) GetGUID() string { return d.ID }
func (d *Document) GetName() string { return d.Title }

func ExampleContainer() {
	c := mkcontainer.New()

	c.Store(
		&Document{ID: "doc-1", Title: "Report"},
		&Document{ID: "doc-2", Title: "Report"},
		&Document{ID: "doc-3", Title: "Summary"},
	)

	// Lookup by unique GUID
	doc := c.GetByGUID("doc-1")

	// Lookup all documents named "Report"
	reports := c.GetByName("Report") // returns 2 items

	fmt.Print(doc.(*Document).ID)
	fmt.Print(doc.(*Document).Title)
	fmt.Print(reports[0].(*Document).ID)
	fmt.Print(reports[0].(*Document).Title)
	fmt.Print(reports[1].(*Document).ID)
	fmt.Print(reports[1].(*Document).Title)
	//output: doc-1Reportdoc-1Reportdoc-2Report
}
