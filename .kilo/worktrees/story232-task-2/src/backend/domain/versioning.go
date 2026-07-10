package domain

import "time"

// Bounded encapsulates temporal validity for an entity. Field names match
// schema column names after Go-Jet normalization so qrm hydrates them via
// the embedding entity's table prefix (set by an `alias` tag on the embed).
type Bounded struct {
	ValidityStart time.Time `json:"validity_start"`
	ValidityEnd   time.Time `json:"validity_end"`
}

// Keyed encapsulates the natural key for an entity. The Key field maps to
// the schema's 'key' column.
type Keyed struct {
	Key string `json:"key"`
}

// Audit encapsulates transaction-level metadata set by the persistence layer.
// Domain code should not modify these fields; they are set by EBM during
// mutations.
type Audit struct {
	TransactionStart time.Time `json:"transaction_start"`
}
