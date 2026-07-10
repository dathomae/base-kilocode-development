package rdbms

// minimalLockable is a deliberately bare implementation of LockableEntity that
// omits MaxTransactionStart. Its sole purpose is to act as a compile-time
// assertion that LockableEntity does not require MaxTransactionStart: the
// staleness check is performed by EBM against an explicitly-passed
// maxTransactionStart parameter, not by interrogating the entity itself.
type minimalLockable struct{}

func (minimalLockable) EntityID() int64    { return 0 }
func (minimalLockable) NaturalKey() string { return "" }

var _ LockableEntity = minimalLockable{}
