package domain

// Namespace is the domain model for the namespace table.
type Namespace struct {
	Keyed             `alias:"namespace"`
	Bounded           `alias:"namespace"`
	Audit             `alias:"namespace"`
	NsID              int64  `json:"ns_id"`
	NsFamID           int32  `json:"ns_fam_id"`
	OwnerSubjectID    int32  `json:"owner_subject_id"`
	NsTypeID          int16  `json:"ns_type_id"`
	InsertionMetadata string `json:"insertion_metadata"`
}

func (n Namespace) EntityID() int64    { return n.NsID }
func (n Namespace) NaturalKey() string { return n.Key }
