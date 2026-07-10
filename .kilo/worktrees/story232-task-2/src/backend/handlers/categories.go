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

type listCategoriesInput struct {
	Family    string `path:"family"    doc:"Namespace family key"`
	Namespace string `path:"namespace" doc:"Namespace key"`
}

type getCategoryInput struct {
	Family      string `path:"family"       doc:"Namespace family key"`
	Namespace   string `path:"namespace"    doc:"Namespace key"`
	CategoryKey string `path:"category_key" doc:"Category key"`
}

type putCategoryInput struct {
	Family      string `path:"family"       doc:"Namespace family key"`
	Namespace   string `path:"namespace"    doc:"Namespace key"`
	CategoryKey string `path:"category_key" doc:"Category key"`
	Body        struct {
		ValidFrom string  `json:"valid_from"         doc:"When this category becomes valid. Use 'now' or RFC3339." example:"2025-01-01T00:00:00Z"`
		ValidTo   *string `json:"valid_to,omitempty" doc:"When this category ceases to be valid. Use 'infinity' or omit for open-ended." example:"2026-01-01T00:00:00Z"`
	}
}

type patchCategoryInput struct {
	Family      string `path:"family"       doc:"Namespace family key"`
	Namespace   string `path:"namespace"    doc:"Namespace key"`
	CategoryKey string `path:"category_key" doc:"Category key"`
	Body        struct {
		ValidFrom *string `json:"valid_from,omitempty" doc:"Update valid_from. Use 'now' or RFC3339. Omit to leave unchanged."`
		ValidTo   *string `json:"valid_to,omitempty"   doc:"Update valid_to. Use 'infinity' for open-ended, RFC3339 for specific end. Omit to leave unchanged."`
	}
}

type categoryOutput struct {
	CategoryID       int64     `json:"category_id"`
	NsID             int64     `json:"ns_id"`
	Key              string    `json:"key"`
	ValidityStart    time.Time `json:"validity_start"`
	ValidityEnd      time.Time `json:"validity_end"`
	TransactionStart time.Time `json:"transaction_start"`
}

type categoriesOutput struct {
	Categories []categoryOutput `json:"categories"`
}

type categoryResponse struct {
	Body categoryOutput
}

type categoriesResponse struct {
	Body categoriesOutput
}

type categoryShapeOutput struct {
	CategoryAttrID   int64     `json:"category_attr_id"`
	CategoryID       int64     `json:"category_id"`
	AttrOwnerNsID    int64     `json:"attr_owner_ns_id"`
	Key              string    `json:"key"`
	Value            string    `json:"value"`
	ValidityStart    time.Time `json:"validity_start"`
	ValidityEnd      time.Time `json:"validity_end"`
	TransactionStart time.Time `json:"transaction_start"`
}

type categoryShapesOutput struct {
	Shapes []categoryShapeOutput `json:"shapes"`
}

type categoryShapeResponse struct {
	Body categoryShapeOutput
}

type categoryShapesResponse struct {
	Body categoryShapesOutput
}

type putCategoryShapeInput struct {
	Family             string `path:"family"              doc:"Namespace family key"`
	Namespace          string `path:"namespace"           doc:"Namespace key"`
	CategoryKey        string `path:"category_key"        doc:"Category key"`
	AssertingFamily    string `path:"asserting_family"    doc:"Asserting namespace family key"`
	AssertingNamespace string `path:"asserting_namespace" doc:"Asserting namespace key"`
	ShapeKey           string `path:"shape_key"           doc:"Attribute key"`
	Body               struct {
		IsRequired bool    `json:"is_required"          doc:"Whether this attribute is mandatory"`
		JSONSchema string  `json:"json_schema"          doc:"JSON Schema for attribute value validation"`
		ValidFrom  string  `json:"valid_from"           doc:"When this shape becomes valid"`
		ValidTo    *string `json:"valid_to,omitempty"   doc:"When this shape ceases to be valid"`
	}
}

func categoryToOutput(c *domain.Category) categoryOutput {
	return categoryOutput{
		CategoryID:       c.CategoryID,
		NsID:             c.NsID,
		Key:              c.Key,
		ValidityStart:    c.ValidityStart,
		ValidityEnd:      c.ValidityEnd,
		TransactionStart: c.TransactionStart,
	}
}

// resolveNamespace looks up family then namespace, returning 404 errors inline.
func resolveNamespace(ctx context.Context, db *sql.DB, familyKey, namespaceKey string) (*domain.Namespace, error) {
	fam, err := domain.GetFamilyByKey(ctx, db, familyKey, nil, nil)
	if err != nil {
		return nil, huma.NewError(http.StatusInternalServerError, fmt.Sprintf("database error: %v", err))
	}
	if fam == nil {
		return nil, huma.NewError(http.StatusNotFound, "namespace family not found")
	}

	ns, err := domain.GetNamespaceByKey(ctx, db, fam.NsFamID, namespaceKey, nil, nil)
	if err != nil {
		return nil, huma.NewError(http.StatusInternalServerError, fmt.Sprintf("database error: %v", err))
	}
	if ns == nil {
		return nil, huma.NewError(http.StatusNotFound, "namespace not found")
	}
	return ns, nil
}

func RegisterCategoryRoutes(api huma.API, db *sql.DB) {
	huma.Register(api, huma.Operation{
		OperationID: "list-categories",
		Method:      http.MethodGet,
		Path:        "/categories/{family}/{namespace}",
		Summary:     "List all categories within a namespace",
	}, func(ctx context.Context, input *listCategoriesInput) (*categoriesResponse, error) {
		ns, err := resolveNamespace(ctx, db, input.Family, input.Namespace)
		if err != nil {
			return nil, err
		}

		categories, err := domain.ListCategories(ctx, db, ns.NsID, nil, nil)
		if err != nil {
			return nil, huma.NewError(http.StatusInternalServerError, fmt.Sprintf("database error: %v", err))
		}

		output := make([]categoryOutput, len(categories))
		for i := range categories {
			output[i] = categoryToOutput(&categories[i])
		}
		return &categoriesResponse{Body: categoriesOutput{Categories: output}}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "get-category",
		Method:      http.MethodGet,
		Path:        "/category/{family}/{namespace}/{category_key}",
		Summary:     "Get a category by full path",
	}, func(ctx context.Context, input *getCategoryInput) (*categoryResponse, error) {
		ns, err := resolveNamespace(ctx, db, input.Family, input.Namespace)
		if err != nil {
			return nil, err
		}

		cat, err := domain.GetCategoryByKey(ctx, db, ns.NsID, input.CategoryKey, nil, nil)
		if err != nil {
			return nil, huma.NewError(http.StatusInternalServerError, fmt.Sprintf("database error: %v", err))
		}
		if cat == nil {
			return nil, huma.NewError(http.StatusNotFound, "category not found")
		}

		return &categoryResponse{Body: categoryToOutput(cat)}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "put-category",
		Method:      http.MethodPut,
		Path:        "/category/{family}/{namespace}/{category_key}",
		Summary:     "Create or replace a category",
	}, func(ctx context.Context, input *putCategoryInput) (*categoryResponse, error) {
		if len(input.CategoryKey) < 6 || len(input.CategoryKey) > 64 {
			return nil, huma.NewError(http.StatusUnprocessableEntity, "key must be 6-64 characters")
		}

		ns, err := resolveNamespace(ctx, db, input.Family, input.Namespace)
		if err != nil {
			return nil, err
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

		cat, err := domain.CreateOrUpdateCategory(ctx, db, ns.NsID, input.CategoryKey, validFrom, validTo, nil, nil)
		if err != nil {
			return nil, huma.NewError(http.StatusInternalServerError, fmt.Sprintf("mutation error: %v", err))
		}

		return &categoryResponse{Body: categoryToOutput(cat)}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "patch-category",
		Method:      http.MethodPatch,
		Path:        "/category/{family}/{namespace}/{category_key}",
		Summary:     "Update validity of an existing category",
	}, func(ctx context.Context, input *patchCategoryInput) (*categoryResponse, error) {
		ns, err := resolveNamespace(ctx, db, input.Family, input.Namespace)
		if err != nil {
			return nil, err
		}

		cat, err := domain.GetCategoryByKey(ctx, db, ns.NsID, input.CategoryKey, nil, nil)
		if err != nil {
			return nil, huma.NewError(http.StatusInternalServerError, fmt.Sprintf("database error: %v", err))
		}
		if cat == nil {
			return nil, huma.NewError(http.StatusNotFound, "category not found")
		}

		if input.Body.ValidFrom != nil {
			if cat.ValidityStart, err = parseTimestamp(*input.Body.ValidFrom); err != nil {
				return nil, huma.NewError(http.StatusUnprocessableEntity, fmt.Sprintf("invalid valid_from: %v", err))
			}
		}
		if input.Body.ValidTo != nil {
			if cat.ValidityEnd, err = parseTimestamp(*input.Body.ValidTo); err != nil {
				return nil, huma.NewError(http.StatusUnprocessableEntity, fmt.Sprintf("invalid valid_to: %v", err))
			}
		}

		updated, err := domain.CreateOrUpdateCategory(ctx, db, ns.NsID, cat.Key, cat.ValidityStart, cat.ValidityEnd, nil, nil)
		if err != nil {
			return nil, huma.NewError(http.StatusInternalServerError, fmt.Sprintf("mutation error: %v", err))
		}

		return &categoryResponse{Body: categoryToOutput(updated)}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "list-category-shapes",
		Method:      http.MethodGet,
		Path:        "/category-shapes/{family}/{namespace}/{category_key}",
		Summary:     "List all attribute shapes for a category",
	}, func(ctx context.Context, input *getCategoryInput) (*categoryShapesResponse, error) {
		return nil, huma.NewError(http.StatusNotImplemented, "category shapes are returned through aggregate category reads in Story 043L")
	})

	huma.Register(api, huma.Operation{
		OperationID: "put-category-shape",
		Method:      http.MethodPut,
		Path:        "/category-shape/{family}/{namespace}/{category_key}/{asserting_family}/{asserting_namespace}/{shape_key}",
		Summary:     "Create or replace an attribute shape",
	}, func(ctx context.Context, input *putCategoryShapeInput) (*categoryShapeResponse, error) {
		return nil, huma.NewError(http.StatusNotImplemented, "category shapes not yet implemented")
	})
}
