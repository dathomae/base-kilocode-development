package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	c "chronocone/backend/constants"
	"chronocone/backend/domain"
	"github.com/danielgtaylor/huma/v2"
)

type listNamespacesInput struct {
	FamilyKey string `path:"family_key" doc:"Namespace family key"`
}

type getNamespaceInput struct {
	FamilyKey    string `path:"family_key"    doc:"Namespace family key"`
	NamespaceKey string `path:"namespace_key" doc:"Namespace key"`
}

type putNamespaceInput struct {
	FamilyKey    string `path:"family_key"    doc:"Namespace family key"`
	NamespaceKey string `path:"namespace_key" doc:"Namespace key"`
	Body         struct {
		ValidFrom string  `json:"valid_from"         doc:"When this namespace becomes valid. Use 'now' or RFC3339." example:"2025-01-01T00:00:00Z"`
		ValidTo   *string `json:"valid_to,omitempty" doc:"When this namespace ceases to be valid. Use 'infinity' or omit for open-ended." example:"2026-01-01T00:00:00Z"`
	}
}

type patchNamespaceInput struct {
	FamilyKey    string `path:"family_key"    doc:"Namespace family key"`
	NamespaceKey string `path:"namespace_key" doc:"Namespace key"`
	Body         struct {
		ValidFrom *string `json:"valid_from,omitempty" doc:"Update valid_from. Use 'now' or RFC3339. Omit to leave unchanged."`
		ValidTo   *string `json:"valid_to,omitempty"   doc:"Update valid_to. Use 'infinity' for open-ended, RFC3339 for specific end. Omit to leave unchanged."`
	}
}

type namespaceOutput struct {
	NsID             int64     `json:"ns_id"`
	NsFamID          int32     `json:"ns_fam_id"`
	Key              string    `json:"key"`
	ValidityStart    time.Time `json:"validity_start"`
	ValidityEnd      time.Time `json:"validity_end"`
	TransactionStart time.Time `json:"transaction_start"`
	OwnerSubjectID   int32     `json:"owner_subject_id"`
	NsTypeID         int16     `json:"ns_type_id"`
}

type namespacesOutput struct {
	Namespaces []namespaceOutput `json:"namespaces"`
}

type namespaceResponse struct {
	Body namespaceOutput
}

type namespacesResponse struct {
	Body namespacesOutput
}

func namespaceToOutput(n *domain.Namespace) namespaceOutput {
	return namespaceOutput{
		NsID:             n.NsID,
		NsFamID:          n.NsFamID,
		Key:              n.Key,
		ValidityStart:    n.ValidityStart,
		ValidityEnd:      n.ValidityEnd,
		TransactionStart: n.TransactionStart,
		OwnerSubjectID:   n.OwnerSubjectID,
		NsTypeID:         n.NsTypeID,
	}
}

func RegisterNamespaceRoutes(api huma.API, db *sql.DB) {
	huma.Register(api, huma.Operation{
		OperationID: "list-namespaces",
		Method:      http.MethodGet,
		Path:        "/namespaces/{family_key}",
		Summary:     "List all namespaces within a family",
	}, func(ctx context.Context, input *listNamespacesInput) (*namespacesResponse, error) {
		fam, err := domain.GetFamilyByKey(ctx, db, input.FamilyKey, nil, nil)
		if err != nil {
			return nil, huma.NewError(http.StatusInternalServerError, fmt.Sprintf("database error: %v", err))
		}
		if fam == nil {
			return nil, huma.NewError(http.StatusNotFound, "namespace family not found")
		}

		namespaces, err := domain.ListNamespaces(ctx, db, fam.NsFamID, nil, nil)
		if err != nil {
			return nil, huma.NewError(http.StatusInternalServerError, fmt.Sprintf("database error: %v", err))
		}

		output := make([]namespaceOutput, len(namespaces))
		for i := range namespaces {
			output[i] = namespaceToOutput(&namespaces[i])
		}
		return &namespacesResponse{Body: namespacesOutput{Namespaces: output}}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "get-namespace",
		Method:      http.MethodGet,
		Path:        "/namespace/{family_key}/{namespace_key}",
		Summary:     "Get a namespace by family and key",
	}, func(ctx context.Context, input *getNamespaceInput) (*namespaceResponse, error) {
		fam, err := domain.GetFamilyByKey(ctx, db, input.FamilyKey, nil, nil)
		if err != nil {
			return nil, huma.NewError(http.StatusInternalServerError, fmt.Sprintf("database error: %v", err))
		}
		if fam == nil {
			return nil, huma.NewError(http.StatusNotFound, "namespace family not found")
		}

		ns, err := domain.GetNamespaceByKey(ctx, db, fam.NsFamID, input.NamespaceKey, nil, nil)
		if err != nil {
			return nil, huma.NewError(http.StatusInternalServerError, fmt.Sprintf("database error: %v", err))
		}
		if ns == nil {
			return nil, huma.NewError(http.StatusNotFound, "namespace not found")
		}

		return &namespaceResponse{Body: namespaceToOutput(ns)}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "put-namespace",
		Method:      http.MethodPut,
		Path:        "/namespace/{family_key}/{namespace_key}",
		Summary:     "Create or replace a namespace",
	}, func(ctx context.Context, input *putNamespaceInput) (*namespaceResponse, error) {
		if len(input.NamespaceKey) < 6 || len(input.NamespaceKey) > 64 {
			return nil, huma.NewError(http.StatusUnprocessableEntity, "key must be 6-64 characters")
		}

		fam, err := domain.GetFamilyByKey(ctx, db, input.FamilyKey, nil, nil)
		if err != nil {
			return nil, huma.NewError(http.StatusInternalServerError, fmt.Sprintf("database error: %v", err))
		}
		if fam == nil {
			return nil, huma.NewError(http.StatusNotFound, "namespace family not found")
		}

		validFrom, err := parseTimestamp(input.Body.ValidFrom)
		if err != nil {
			return nil, huma.NewError(http.StatusUnprocessableEntity, fmt.Sprintf("invalid valid_from: %v", err))
		}

		validTo := c.InfinityTime
		if input.Body.ValidTo != nil {
			if validTo, err = parseTimestamp(*input.Body.ValidTo); err != nil {
				return nil, huma.NewError(http.StatusUnprocessableEntity, fmt.Sprintf("invalid valid_to: %v", err))
			}
		}

		ns, err := domain.CreateOrUpdateNamespace(ctx, db, fam.NsFamID, input.NamespaceKey, validFrom, validTo, nil, nil)
		if err != nil {
			return nil, huma.NewError(http.StatusInternalServerError, fmt.Sprintf("mutation error: %v", err))
		}

		return &namespaceResponse{Body: namespaceToOutput(ns)}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "patch-namespace",
		Method:      http.MethodPatch,
		Path:        "/namespace/{family_key}/{namespace_key}",
		Summary:     "Update validity of an existing namespace",
	}, func(ctx context.Context, input *patchNamespaceInput) (*namespaceResponse, error) {
		fam, err := domain.GetFamilyByKey(ctx, db, input.FamilyKey, nil, nil)
		if err != nil {
			return nil, huma.NewError(http.StatusInternalServerError, fmt.Sprintf("database error: %v", err))
		}
		if fam == nil {
			return nil, huma.NewError(http.StatusNotFound, "namespace family not found")
		}

		ns, err := domain.GetNamespaceByKey(ctx, db, fam.NsFamID, input.NamespaceKey, nil, nil)
		if err != nil {
			return nil, huma.NewError(http.StatusInternalServerError, fmt.Sprintf("database error: %v", err))
		}
		if ns == nil {
			return nil, huma.NewError(http.StatusNotFound, "namespace not found")
		}

		if input.Body.ValidFrom != nil {
			if ns.ValidityStart, err = parseTimestamp(*input.Body.ValidFrom); err != nil {
				return nil, huma.NewError(http.StatusUnprocessableEntity, fmt.Sprintf("invalid valid_from: %v", err))
			}
		}
		if input.Body.ValidTo != nil {
			if ns.ValidityEnd, err = parseTimestamp(*input.Body.ValidTo); err != nil {
				return nil, huma.NewError(http.StatusUnprocessableEntity, fmt.Sprintf("invalid valid_to: %v", err))
			}
		}

		updated, err := domain.CreateOrUpdateNamespace(ctx, db, fam.NsFamID, ns.Key, ns.ValidityStart, ns.ValidityEnd, nil, nil)
		if err != nil {
			return nil, huma.NewError(http.StatusInternalServerError, fmt.Sprintf("mutation error: %v", err))
		}

		return &namespaceResponse{Body: namespaceToOutput(updated)}, nil
	})
}
