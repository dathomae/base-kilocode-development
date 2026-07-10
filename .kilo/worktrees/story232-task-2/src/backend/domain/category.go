package domain

// Category is the domain model for the category table.
//
// Go-Jet alignment: Go type name normalizes to the table name, direct field
// names normalize to column names, embedded persistence mixins carry an
// `alias:"category"` tag.
type Category struct {
	Keyed             `alias:"category"`
	Bounded           `alias:"category"`
	Audit             `alias:"category"`
	CategoryID        int64  `json:"category_id"`
	NsID              int64  `json:"ns_id"`
	InsertionMetadata string `json:"insertion_metadata"`
}

func (c Category) EntityID() int64    { return c.CategoryID }
func (c Category) NaturalKey() string { return c.Key }
