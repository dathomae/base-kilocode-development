package domain

// Element is the domain model for the element table.
type Element struct {
	Keyed             `alias:"element"`
	Bounded           `alias:"element"`
	Audit             `alias:"element"`
	ElementID         int64  `json:"element_id"`
	CategoryID        int64  `json:"category_id"`
	InsertionMetadata string `json:"insertion_metadata"`
}

func (e Element) EntityID() int64    { return e.ElementID }
func (e Element) NaturalKey() string { return e.Key }
