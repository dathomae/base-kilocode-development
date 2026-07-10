package domain

// CategoryShape is the domain model for the category_shape table.
type CategoryShape struct {
	Keyed             `alias:"category_shape"`
	Bounded           `alias:"category_shape"`
	Audit             `alias:"category_shape"`
	CategoryShapeID   int64  `json:"category_shape_id"`
	CategoryID        int64  `json:"category_id"`
	AssertingNsID     int64  `json:"asserting_ns_id"`
	IsRequired        bool   `json:"is_required"`
	JSONSchema        string `json:"json_schema"`
	InsertionMetadata string `json:"insertion_metadata"`
}

func (c CategoryShape) EntityID() int64    { return c.CategoryShapeID }
func (c CategoryShape) NaturalKey() string { return c.Key }

// CategoryAttr is the domain model for the category_attr table.
type CategoryAttr struct {
	Keyed             `alias:"category_attr"`
	Bounded           `alias:"category_attr"`
	Audit             `alias:"category_attr"`
	CategoryAttrID    int64  `json:"category_attr_id"`
	CategoryID        int64  `json:"category_id"`
	AttrOwnerNsID     int64  `json:"attr_owner_ns_id"`
	Value             string `json:"value"`
	InsertionMetadata string `json:"insertion_metadata"`
}

func (c CategoryAttr) EntityID() int64    { return c.CategoryAttrID }
func (c CategoryAttr) NaturalKey() string { return c.Key }
