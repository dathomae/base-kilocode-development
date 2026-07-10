package handlers

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Example: A planner tracks individual line items (like specific budget codes or headcount slots)
func TestElementHandlers(t *testing.T) {
	_, router := createTestAPI(t, testDB)

	t.Run("Example_CreateElementInHierarchy", func(t *testing.T) {
		// SCENARIO: The planner creates a specific budget code within the Budget category
		// BEFORE: Family → Namespace → Category exist, but no elements

		doRequest(t, router, "PUT", "/family/acmeco", map[string]interface{}{"valid_from": "now"})
		doRequest(t, router, "PUT", "/namespace/acmeco/finance", map[string]interface{}{"valid_from": "now"})
		doRequest(t, router, "PUT", "/category/acmeco/finance/budget", map[string]interface{}{"valid_from": "now"})

		// Create a specific budget code (an element)
		createResp, createBody := doRequest(t, router, "PUT", "/element/acmeco/finance/budget/operatingbudget2025", map[string]interface{}{
			"valid_from": "now",
		})

		// AFTER: The element is created
		assert.Equal(t, http.StatusOK, createResp.StatusCode)
		var created elementOutput
		require.NoError(t, json.Unmarshal(createBody, &created))
		assert.Equal(t, "operatingbudget2025", created.Key)

		// The planner can retrieve it
		getResp, getBody := doRequest(t, router, "GET", "/element/acmeco/finance/budget/operatingbudget2025", nil)
		assert.Equal(t, http.StatusOK, getResp.StatusCode)

		var retrieved elementOutput
		require.NoError(t, json.Unmarshal(getBody, &retrieved))
		assert.Equal(t, "operatingbudget2025", retrieved.Key)
	})

	t.Run("Example_ListElementsInCategory", func(t *testing.T) {
		// SCENARIO: The planner has created multiple budget codes and wants to review them
		doRequest(t, router, "PUT", "/family/corpco", map[string]interface{}{"valid_from": "now"})
		doRequest(t, router, "PUT", "/namespace/corpco/finance", map[string]interface{}{"valid_from": "now"})
		doRequest(t, router, "PUT", "/category/corpco/finance/budgets", map[string]interface{}{"valid_from": "now"})

		// Create several budget elements
		doRequest(t, router, "PUT", "/element/corpco/finance/budgets/capexco", map[string]interface{}{"valid_from": "now"})
		doRequest(t, router, "PUT", "/element/corpco/finance/budgets/opexco", map[string]interface{}{"valid_from": "now"})
		doRequest(t, router, "PUT", "/element/corpco/finance/budgets/marketing", map[string]interface{}{"valid_from": "now"})
		doRequest(t, router, "PUT", "/element/corpco/finance/budgets/rdteam", map[string]interface{}{"valid_from": "now"})

		// List all budget elements
		resp, body := doRequest(t, router, "GET", "/elements/corpco/finance/budgets", nil)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result elementsOutput
		require.NoError(t, json.Unmarshal(body, &result))
		assert.GreaterOrEqual(t, len(result.Elements), 4)
	})

	t.Run("Example_ElementMustHaveValidParent", func(t *testing.T) {
		// SCENARIO: The planner tries to create an element in a category that doesn't exist
		resp, body := doRequest(t, router, "PUT", "/element/nonexist/nonexist/nonexist/budgetcode", map[string]interface{}{
			"valid_from": "now",
		})

		// The API rejects it because the parent category doesn't exist
		assert.Equal(t, http.StatusNotFound, resp.StatusCode, "response: %s", string(body))
	})

	t.Run("Example_TrackLineItemsWithinCategory", func(t *testing.T) {
		// SCENARIO: The planner tracks multiple headcount slots within an Org Headcount category
		doRequest(t, router, "PUT", "/family/enterprise", map[string]interface{}{"valid_from": "now"})
		doRequest(t, router, "PUT", "/namespace/enterprise/hrteam", map[string]interface{}{"valid_from": "now"})
		doRequest(t, router, "PUT", "/category/enterprise/hrteam/headcount", map[string]interface{}{"valid_from": "now"})

		// Create headcount elements for different departments
		doRequest(t, router, "PUT", "/element/enterprise/hrteam/headcount/engineering", map[string]interface{}{"valid_from": "now"})
		doRequest(t, router, "PUT", "/element/enterprise/hrteam/headcount/salesteam", map[string]interface{}{"valid_from": "now"})
		doRequest(t, router, "PUT", "/element/enterprise/hrteam/headcount/operations", map[string]interface{}{"valid_from": "now"})

		// Retrieve a specific headcount allocation
		resp, body := doRequest(t, router, "GET", "/element/enterprise/hrteam/headcount/engineering", nil)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result elementOutput
		require.NoError(t, json.Unmarshal(body, &result))
		assert.Equal(t, "engineering", result.Key)
	})
}
