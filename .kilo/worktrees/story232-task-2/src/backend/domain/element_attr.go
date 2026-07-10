package domain

import "strconv"

// ElementAttr is the domain model for the element_attr table. The table has
// no 'key' column — its identity is fully captured by element_attr_id (a
// surrogate). NaturalKey returns the surrogate as a string so EBM's
// post-mutation row-matching can correlate input and fetched rows.
type ElementAttr struct {
	Bounded           `alias:"element_attr"`
	Audit             `alias:"element_attr"`
	ElementAttrID     int64  `json:"element_attr_id"`
	ElementID         int64  `json:"element_id"`
	CategoryShapeID   int64  `json:"category_shape_id"`
	Value             string `json:"value"`
	InsertionMetadata string `json:"insertion_metadata"`
}

func (e ElementAttr) EntityID() int64    { return e.ElementAttrID }
func (e ElementAttr) NaturalKey() string { return strconv.FormatInt(e.ElementAttrID, 10) }
