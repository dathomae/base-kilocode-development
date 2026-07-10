package handlers

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Example: A planner organizes planning line items within a namespace
func TestCategoryHandlers(t *testing.T) {
	_, router := createTestAPI(t, testDB)

	t.Run("Example_CreateAndRetrieveCategory", func(t *testing.T) {
		// SCENARIO: The planner creates planning categories (Budget, Headcount) within Finance
		// BEFORE: Family and namespace exist, but no categories

		doRequest(t, router, "PUT", "/family/corpco", map[string]interface{}{"valid_from": "now"})
		doRequest(t, router, "PUT", "/namespace/corpco/finance", map[string]interface{}{"valid_from": "now"})

		// Create the Budget category
		createResp, createBody := doRequest(t, router, "PUT", "/category/corpco/finance/budget", map[string]interface{}{
			"valid_from": "now",
		})

		// AFTER: The category is created
		assert.Equal(t, http.StatusOK, createResp.StatusCode)
		var created categoryOutput
		require.NoError(t, json.Unmarshal(createBody, &created))
		assert.Equal(t, "budget", created.Key)

		// The planner can retrieve it
		getResp, getBody := doRequest(t, router, "GET", "/category/corpco/finance/budget", nil)
		assert.Equal(t, http.StatusOK, getResp.StatusCode)

		var retrieved categoryOutput
		require.NoError(t, json.Unmarshal(getBody, &retrieved))
		assert.Equal(t, "budget", retrieved.Key)
	})

	t.Run("Example_ListCategoriesInNamespace", func(t *testing.T) {
		// SCENARIO: The planner created multiple budget categories and wants to see them all
		doRequest(t, router, "PUT", "/family/enterprise", map[string]interface{}{"valid_from": "now"})
		doRequest(t, router, "PUT", "/namespace/enterprise/finance", map[string]interface{}{"valid_from": "now"})
		doRequest(t, router, "PUT", "/category/enterprise/finance/capitalbudget", map[string]interface{}{"valid_from": "now"})
		doRequest(t, router, "PUT", "/category/enterprise/finance/operatingbudget", map[string]interface{}{"valid_from": "now"})
		doRequest(t, router, "PUT", "/category/enterprise/finance/cashflow", map[string]interface{}{"valid_from": "now"})

		// List all categories in Finance
		resp, body := doRequest(t, router, "GET", "/categories/enterprise/finance", nil)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result categoriesOutput
		require.NoError(t, json.Unmarshal(body, &result))
		assert.GreaterOrEqual(t, len(result.Categories), 3)
	})

	t.Run("Example_CategoryMustHaveValidParent", func(t *testing.T) {
		// SCENARIO: The planner tries to create a category in a namespace that doesn't exist
		resp, body := doRequest(t, router, "PUT", "/category/nonexist/nonexist/budget", map[string]interface{}{
			"valid_from": "now",
		})

		// The API rejects it because the parent namespace doesn't exist
		assert.Equal(t, http.StatusNotFound, resp.StatusCode, "response: %s", string(body))
	})

	t.Run("Example_RetrieveSpecificCategory", func(t *testing.T) {
		// SCENARIO: The planner wants to check details of a specific planning category
		doRequest(t, router, "PUT", "/family/retail", map[string]interface{}{"valid_from": "now"})
		doRequest(t, router, "PUT", "/namespace/retail/operations", map[string]interface{}{"valid_from": "now"})
		doRequest(t, router, "PUT", "/category/retail/operations/storemap", map[string]interface{}{"valid_from": "now"})

		// Get the specific category
		resp, body := doRequest(t, router, "GET", "/category/retail/operations/storemap", nil)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result categoryOutput
		require.NoError(t, json.Unmarshal(body, &result))
		assert.Equal(t, "storemap", result.Key)
	})
}
