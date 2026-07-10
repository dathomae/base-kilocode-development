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

type getFamilyInput struct {
	FamilyKey string `path:"family_key" doc:"Namespace family key"`
}

type putFamilyInput struct {
	FamilyKey string `path:"family_key" doc:"Namespace family key"`
	Body      struct {
		ValidFrom string  `json:"valid_from"         doc:"When this family becomes valid. Use 'now' or RFC3339." example:"2025-01-01T00:00:00Z"`
		ValidTo   *string `json:"valid_to,omitempty" doc:"When this family ceases to be valid. Use 'infinity' or omit for open-ended." example:"2026-01-01T00:00:00Z"`
	}
}

type patchFamilyInput struct {
	FamilyKey string `path:"family_key" doc:"Namespace family key"`
	Body      struct {
		ValidFrom *string `json:"valid_from,omitempty" doc:"Update valid_from. Use 'now' or RFC3339. Omit to leave unchanged."`
		ValidTo   *string `json:"valid_to,omitempty"   doc:"Update valid_to. Use 'infinity' for open-ended, RFC3339 for specific end. Omit to leave unchanged."`
	}
}

type familyOutput struct {
	NsFamID           int32     `json:"ns_fam_id"`
	Key               string    `json:"key"`
	ValidityStart     time.Time `json:"validity_start"`
	ValidityEnd       time.Time `json:"validity_end"`
	TransactionStart  time.Time `json:"transaction_start"`
	CustomerSubjectID int32     `json:"customer_subject_id"`
	OwnerSubjectID    int32     `json:"owner_subject_id"`
	NsFamTypeID       int16     `json:"ns_fam_type_id"`
}

type familiesOutput struct {
	Families []familyOutput `json:"families"`
}

type familyResponse struct {
	Body familyOutput
}

type familiesResponse struct {
	Body familiesOutput
}

func familyToOutput(f *domain.NamespaceFamily) familyOutput {
	return familyOutput{
		NsFamID:           f.NsFamID,
		Key:               f.Key,
		ValidityStart:     f.ValidityStart,
		ValidityEnd:       f.ValidityEnd,
		TransactionStart:  f.TransactionStart,
		CustomerSubjectID: f.CustomerSubjectID,
		OwnerSubjectID:    f.OwnerSubjectID,
		NsFamTypeID:       f.NsFamTypeID,
	}
}

func RegisterFamilyRoutes(api huma.API, db *sql.DB) {
	huma.Register(api, huma.Operation{
		OperationID: "list-families",
		Method:      http.MethodGet,
		Path:        "/families",
		Summary:     "List all namespace families",
	}, func(ctx context.Context, _ *struct{}) (*familiesResponse, error) {
		families, err := domain.ListFamilies(ctx, db, nil, nil)
		if err != nil {
			return nil, huma.NewError(http.StatusInternalServerError, fmt.Sprintf("database error: %v", err))
		}

		output := make([]familyOutput, len(families))
		for i := range families {
			output[i] = familyToOutput(&families[i])
		}
		return &familiesResponse{Body: familiesOutput{Families: output}}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "get-family",
		Method:      http.MethodGet,
		Path:        "/family/{family_key}",
		Summary:     "Get a namespace family by key",
	}, func(ctx context.Context, input *getFamilyInput) (*familyResponse, error) {
		if len(input.FamilyKey) < 6 || len(input.FamilyKey) > 64 {
			return nil, huma.NewError(http.StatusUnprocessableEntity, "key must be 6-64 characters")
		}

		family, err := domain.GetFamilyByKey(ctx, db, input.FamilyKey, nil, nil)
		if err != nil {
			return nil, huma.NewError(http.StatusInternalServerError, fmt.Sprintf("database error: %v", err))
		}
		if family == nil {
			return nil, huma.NewError(http.StatusNotFound, "namespace family not found")
		}

		return &familyResponse{Body: familyToOutput(family)}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "put-family",
		Method:      http.MethodPut,
		Path:        "/family/{family_key}",
		Summary:     "Create or replace a namespace family",
	}, func(ctx context.Context, input *putFamilyInput) (*familyResponse, error) {
		if len(input.FamilyKey) < 6 || len(input.FamilyKey) > 64 {
			return nil, huma.NewError(http.StatusUnprocessableEntity, "key must be 6-64 characters")
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

		family, err := domain.CreateOrUpdateFamily(ctx, db, input.FamilyKey, validFrom, validTo, nil, nil)
		if err != nil {
			return nil, huma.NewError(http.StatusInternalServerError, fmt.Sprintf("mutation error: %v", err))
		}

		return &familyResponse{Body: familyToOutput(family)}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "patch-family",
		Method:      http.MethodPatch,
		Path:        "/family/{family_key}",
		Summary:     "Update validity of an existing namespace family",
	}, func(ctx context.Context, input *patchFamilyInput) (*familyResponse, error) {
		if len(input.FamilyKey) < 6 || len(input.FamilyKey) > 64 {
			return nil, huma.NewError(http.StatusUnprocessableEntity, "key must be 6-64 characters")
		}

		family, err := domain.GetFamilyByKey(ctx, db, input.FamilyKey, nil, nil)
		if err != nil {
			return nil, huma.NewError(http.StatusInternalServerError, fmt.Sprintf("database error: %v", err))
		}
		if family == nil {
			return nil, huma.NewError(http.StatusNotFound, "namespace family not found")
		}

		if input.Body.ValidFrom != nil {
			if family.ValidityStart, err = parseTimestamp(*input.Body.ValidFrom); err != nil {
				return nil, huma.NewError(http.StatusUnprocessableEntity, fmt.Sprintf("invalid valid_from: %v", err))
			}
		}
		if input.Body.ValidTo != nil {
			if family.ValidityEnd, err = parseTimestamp(*input.Body.ValidTo); err != nil {
				return nil, huma.NewError(http.StatusUnprocessableEntity, fmt.Sprintf("invalid valid_to: %v", err))
			}
		}

		updated, err := domain.CreateOrUpdateFamily(ctx, db, family.Key, family.ValidityStart, family.ValidityEnd, nil, nil)
		if err != nil {
			return nil, huma.NewError(http.StatusInternalServerError, fmt.Sprintf("mutation error: %v", err))
		}

		return &familyResponse{Body: familyToOutput(updated)}, nil
	})
}
