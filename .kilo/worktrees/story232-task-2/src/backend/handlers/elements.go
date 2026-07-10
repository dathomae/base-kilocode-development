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

type listElementsInput struct {
	Family      string `path:"family"       doc:"Namespace family key"`
	Namespace   string `path:"namespace"    doc:"Namespace key"`
	CategoryKey string `path:"category_key" doc:"Category key"`
}

type getElementInput struct {
	Family      string `path:"family"       doc:"Namespace family key"`
	Namespace   string `path:"namespace"    doc:"Namespace key"`
	CategoryKey string `path:"category_key" doc:"Category key"`
	ElementKey  string `path:"element_key"  doc:"Element key"`
}

type putElementInput struct {
	Family      string `path:"family"       doc:"Namespace family key"`
	Namespace   string `path:"namespace"    doc:"Namespace key"`
	CategoryKey string `path:"category_key" doc:"Category key"`
	ElementKey  string `path:"element_key"  doc:"Element key"`
	Body        struct {
		ValidFrom string  `json:"valid_from"         doc:"When this element becomes valid. Use 'now' or RFC3339." example:"2025-01-01T00:00:00Z"`
		ValidTo   *string `json:"valid_to,omitempty" doc:"When this element ceases to be valid. Use 'infinity' or omit for open-ended." example:"2026-01-01T00:00:00Z"`
	}
}

type patchElementInput struct {
	Family      string `path:"family"       doc:"Namespace family key"`
	Namespace   string `path:"namespace"    doc:"Namespace key"`
	CategoryKey string `path:"category_key" doc:"Category key"`
	ElementKey  string `path:"element_key"  doc:"Element key"`
	Body        struct {
		ValidFrom *string `json:"valid_from,omitempty" doc:"Update valid_from. Use 'now' or RFC3339. Omit to leave unchanged."`
		ValidTo   *string `json:"valid_to,omitempty"   doc:"Update valid_to. Use 'infinity' for open-ended, RFC3339 for specific end. Omit to leave unchanged."`
	}
}

type elementOutput struct {
	ElementID        int64     `json:"element_id"`
	CategoryID       int64     `json:"category_id"`
	Key              string    `json:"key"`
	ValidityStart    time.Time `json:"validity_start"`
	ValidityEnd      time.Time `json:"validity_end"`
	TransactionStart time.Time `json:"transaction_start"`
}

type elementsOutput struct {
	Elements []elementOutput `json:"elements"`
}

type elementResponse struct {
	Body elementOutput
}

type elementsResponse struct {
	Body elementsOutput
}

type attributeOutput struct {
	ElementAttrID    int64     `json:"element_attr_id"`
	ElementID        int64     `json:"element_id"`
	CategoryShapeID  int64     `json:"category_shape_id"`
	Value            string    `json:"value"`
	ValidityStart    time.Time `json:"validity_start"`
	ValidityEnd      time.Time `json:"validity_end"`
	TransactionStart time.Time `json:"transaction_start"`
}

type attributesOutput struct {
	Attributes []attributeOutput `json:"attributes"`
}

type attributeResponse struct {
	Body attributeOutput
}

type attributesResponse struct {
	Body attributesOutput
}

type putAttributeInput struct {
	Family             string `path:"family"              doc:"Namespace family key"`
	Namespace          string `path:"namespace"           doc:"Namespace key"`
	CategoryKey        string `path:"category_key"        doc:"Category key"`
	ElementKey         string `path:"element_key"         doc:"Element key"`
	AssertingFamily    string `path:"asserting_family"    doc:"Asserting namespace family key"`
	AssertingNamespace string `path:"asserting_namespace" doc:"Asserting namespace key"`
	ShapeKey           string `path:"shape_key"           doc:"Attribute key"`
	Body               struct {
		Value     string  `json:"value"              doc:"Attribute value (JSON string)"`
		ValidFrom string  `json:"valid_from"         doc:"When this attribute becomes valid"`
		ValidTo   *string `json:"valid_to,omitempty" doc:"When this attribute ceases to be valid"`
	}
}

func elementToOutput(e *domain.Element) elementOutput {
	return elementOutput{
		ElementID:        e.ElementID,
		CategoryID:       e.CategoryID,
		Key:              e.Key,
		ValidityStart:    e.ValidityStart,
		ValidityEnd:      e.ValidityEnd,
		TransactionStart: e.TransactionStart,
	}
}

// resolveCategory looks up family, namespace, and category, returning 404 errors inline.
func resolveCategory(ctx context.Context, db *sql.DB, familyKey, namespaceKey, categoryKey string) (*domain.Category, error) {
	ns, err := resolveNamespace(ctx, db, familyKey, namespaceKey)
	if err != nil {
		return nil, err
	}

	cat, err := domain.GetCategoryByKey(ctx, db, ns.NsID, categoryKey, nil, nil)
	if err != nil {
		return nil, huma.NewError(http.StatusInternalServerError, fmt.Sprintf("database error: %v", err))
	}
	if cat == nil {
		return nil, huma.NewError(http.StatusNotFound, "category not found")
	}
	return cat, nil
}

func RegisterElementRoutes(api huma.API, db *sql.DB) {
	huma.Register(api, huma.Operation{
		OperationID: "list-elements",
		Method:      http.MethodGet,
		Path:        "/elements/{family}/{namespace}/{category_key}",
		Summary:     "List all elements within a category",
	}, func(ctx context.Context, input *listElementsInput) (*elementsResponse, error) {
		cat, err := resolveCategory(ctx, db, input.Family, input.Namespace, input.CategoryKey)
		if err != nil {
			return nil, err
		}

		elements, err := domain.ListElements(ctx, db, cat.CategoryID, nil, nil)
		if err != nil {
			return nil, huma.NewError(http.StatusInternalServerError, fmt.Sprintf("database error: %v", err))
		}

		output := make([]elementOutput, len(elements))
		for i := range elements {
			output[i] = elementToOutput(&elements[i])
		}
		return &elementsResponse{Body: elementsOutput{Elements: output}}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "get-element",
		Method:      http.MethodGet,
		Path:        "/element/{family}/{namespace}/{category_key}/{element_key}",
		Summary:     "Get an element by full path",
	}, func(ctx context.Context, input *getElementInput) (*elementResponse, error) {
		cat, err := resolveCategory(ctx, db, input.Family, input.Namespace, input.CategoryKey)
		if err != nil {
			return nil, err
		}

		elem, err := domain.GetElementByKey(ctx, db, cat.CategoryID, input.ElementKey, nil, nil)
		if err != nil {
			return nil, huma.NewError(http.StatusInternalServerError, fmt.Sprintf("database error: %v", err))
		}
		if elem == nil {
			return nil, huma.NewError(http.StatusNotFound, "element not found")
		}

		return &elementResponse{Body: elementToOutput(elem)}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "put-element",
		Method:      http.MethodPut,
		Path:        "/element/{family}/{namespace}/{category_key}/{element_key}",
		Summary:     "Create or replace an element",
	}, func(ctx context.Context, input *putElementInput) (*elementResponse, error) {
		if len(input.ElementKey) < 6 || len(input.ElementKey) > 64 {
			return nil, huma.NewError(http.StatusUnprocessableEntity, "key must be 6-64 characters")
		}

		cat, err := resolveCategory(ctx, db, input.Family, input.Namespace, input.CategoryKey)
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

		elem, err := domain.CreateOrUpdateElement(ctx, db, cat.CategoryID, input.ElementKey, validFrom, validTo, nil, nil)
		if err != nil {
			return nil, huma.NewError(http.StatusInternalServerError, fmt.Sprintf("mutation error: %v", err))
		}

		return &elementResponse{Body: elementToOutput(elem)}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "patch-element",
		Method:      http.MethodPatch,
		Path:        "/element/{family}/{namespace}/{category_key}/{element_key}",
		Summary:     "Update validity of an existing element",
	}, func(ctx context.Context, input *patchElementInput) (*elementResponse, error) {
		cat, err := resolveCategory(ctx, db, input.Family, input.Namespace, input.CategoryKey)
		if err != nil {
			return nil, err
		}

		elem, err := domain.GetElementByKey(ctx, db, cat.CategoryID, input.ElementKey, nil, nil)
		if err != nil {
			return nil, huma.NewError(http.StatusInternalServerError, fmt.Sprintf("database error: %v", err))
		}
		if elem == nil {
			return nil, huma.NewError(http.StatusNotFound, "element not found")
		}

		if input.Body.ValidFrom != nil {
			if elem.ValidityStart, err = parseTimestamp(*input.Body.ValidFrom); err != nil {
				return nil, huma.NewError(http.StatusUnprocessableEntity, fmt.Sprintf("invalid valid_from: %v", err))
			}
		}
		if input.Body.ValidTo != nil {
			if elem.ValidityEnd, err = parseTimestamp(*input.Body.ValidTo); err != nil {
				return nil, huma.NewError(http.StatusUnprocessableEntity, fmt.Sprintf("invalid valid_to: %v", err))
			}
		}

		updated, err := domain.CreateOrUpdateElement(ctx, db, cat.CategoryID, elem.Key, elem.ValidityStart, elem.ValidityEnd, nil, nil)
		if err != nil {
			return nil, huma.NewError(http.StatusInternalServerError, fmt.Sprintf("mutation error: %v", err))
		}

		return &elementResponse{Body: elementToOutput(updated)}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "list-attributes",
		Method:      http.MethodGet,
		Path:        "/attributes/{family}/{namespace}/{category_key}/{element_key}",
		Summary:     "List all attributes for an element",
	}, func(ctx context.Context, input *getElementInput) (*attributesResponse, error) {
		return nil, huma.NewError(http.StatusNotImplemented, "element attributes are returned through aggregate element reads in Story 043L")
	})

	huma.Register(api, huma.Operation{
		OperationID: "put-attribute",
		Method:      http.MethodPut,
		Path:        "/attribute/{family}/{namespace}/{category_key}/{element_key}/{asserting_family}/{asserting_namespace}/{shape_key}",
		Summary:     "Create or replace an element attribute",
	}, func(ctx context.Context, input *putAttributeInput) (*attributeResponse, error) {
		return nil, huma.NewError(http.StatusNotImplemented, "element attributes not yet implemented")
	})
}
