# OpenAPI Usage Pattern (Go-Native)

This document outlines the patterns and guidelines for using OpenAPI in the Go-native ChronoCone backend.

## Overview

In the Go-native backend, the **Unified Domain Model** (the hand-authored Go structs) acts as the primary source of truth for the API contract. We use **Go-native handlers** and **Middleware** to serve the API, and documentation is maintained through a combination of code-first tagging and manual spec refinement.

## Unified Model as Contract

We avoid separate DTOs for request and response bodies. Instead, the same struct that maps to the database also defines the JSON wire format.

### Field Documentation with JSON Tags

1.  **Snake Case**: All API properties must use `snake_case`.
2.  **Metadata Preservation**: Include bitemporal metadata (like `transaction_start`) in responses to enable staleness detection in the Flush/Purge/Assert mutation protocol.
3.  **Read-Only/Generated Fields**: Use `json:"id,omitempty"` for fields that the client should not provide during creation.
4.  **Sensitive Information**: Use `json:"-"` to prevent internal fields from ever leaking into the API.

```go
type Namespace struct {
    // Human-readable key (Natural Key)
    Name        string    `sql:"name" json:"name"`
    
    // Internal Metadata (Shared for Staleness Detection)
    TransactionStart time.Time `sql:"transaction_start" json:"transaction_start"`
    
    // Database-generated ID (Optional in requests, present in responses)
    ID          int64     `sql:"ns_id" json:"id,omitempty"`
}
```

## API Documentation Strategy

### 1. Manual OpenAPI Specification
The canonical OpenAPI specification is maintained in `resources/apis/openapi/openapi.yaml`. Changes to the **Unified Model** must be reflected in this YAML file to ensure the documentation stays in sync with the code.

### 2. Handler Documentation
Handlers should include comments that describe their behavior, following the **Swaggo/swag** convention where applicable, to allow for automated documentation generation.

```go
// CreateNamespace handles POST /namespaces
// @Summary Create a new namespace
// @Description Create a namespace within a family using the bitemporal protocol.
// @Accept  json
// @Produce  json
// @Param   namespace body domain.Namespace true "Namespace Object"
// @Success 201 {object} domain.Namespace
// @Router /namespaces [post]
func (h *NamespaceHandler) Create(w http.ResponseWriter, r *http.Request) {
    // ...
}
```

## Validation Checklist

When implementing Go-native API endpoints:

- [ ] Does the handler unmarshal directly into the **Unified Domain Model**?
- [ ] Are the `json` tags on the domain struct consistent with the `openapi.yaml`?
- [ ] Does the response include necessary bitemporal metadata for the "round trip" mutation protocol?
- [ ] Are internal-only fields correctly marked with `json:"-"`?
- [ ] Does the handler follow Go's standard `(T, error)` pattern for service calls?
