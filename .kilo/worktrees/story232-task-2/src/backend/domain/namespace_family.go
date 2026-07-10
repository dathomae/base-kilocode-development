package domain

// NamespaceFamily is the domain model for the namespace_family table — the
// root organizational unit for planning hierarchies.
//
// Naming discipline: the Go type name matches the schema table name and the
// direct field names normalize (lowercase + strip underscores) to their
// column names. Embedded persistence mixins carry an `alias:"namespace_family"`
// tag so Go-Jet's qrm rebinds the column-prefix lookup to this table when
// hydrating mixin fields.
type NamespaceFamily struct {
	Keyed             `alias:"namespace_family"`
	Bounded           `alias:"namespace_family"`
	Audit             `alias:"namespace_family"`
	NsFamID           int32  `json:"ns_fam_id"`
	NsFamTypeID       int16  `json:"ns_fam_type_id"`
	CustomerSubjectID int32  `json:"customer_subject_id"`
	OwnerSubjectID    int32  `json:"owner_subject_id"`
	InsertionMetadata string `json:"insertion_metadata"`
}

// EntityID satisfies rdbms.LockableEntity.
func (n NamespaceFamily) EntityID() int64 { return int64(n.NsFamID) }

// NaturalKey satisfies rdbms.LockableEntity. The natural key for a namespace
// family is the value of its 'key' column.
func (n NamespaceFamily) NaturalKey() string { return n.Key }
