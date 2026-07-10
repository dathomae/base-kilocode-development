package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"chronocone/backend/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Example: A planner organizes planning domains within a family
func TestNamespaceHandlers(t *testing.T) {
	_, router := createTestAPI(t, testDB)
	ctx := context.Background()

	t.Run("Example_CreateNamespacesWithinFamily", func(t *testing.T) {
		// SCENARIO: The planner creates functional domains (Finance, Operations) within ACME Corp
		// BEFORE: ACME Corp family exists but has no namespaces

		doRequest(t, router, "PUT", "/family/nsacme", map[string]interface{}{
			"valid_from": "now",
		})

		// Create the Finance namespace under ACME
		createResp, createBody := doRequest(t, router, "PUT", "/namespace/nsacme/finance", map[string]interface{}{
			"valid_from": "now",
		})

		// AFTER: Finance namespace is created and returned
		assert.Equal(t, http.StatusOK, createResp.StatusCode)
		var created namespaceOutput
		require.NoError(t, json.Unmarshal(createBody, &created))
		assert.Equal(t, "finance", created.Key)

		// The planner can retrieve it
		getResp, getBody := doRequest(t, router, "GET", "/namespace/nsacme/finance", nil)
		assert.Equal(t, http.StatusOK, getResp.StatusCode)

		var retrieved namespaceOutput
		require.NoError(t, json.Unmarshal(getBody, &retrieved))
		assert.Equal(t, "finance", retrieved.Key)

		// Verify it persisted to the database via the domain layer
		fam, err := domain.GetFamilyByKey(ctx, testDB, "nsacme", nil, nil)
		require.NoError(t, err)
		require.NotNil(t, fam)

		ns, err := domain.GetNamespaceByKey(ctx, testDB, fam.NsFamID, "finance", nil, nil)
		require.NoError(t, err)
		require.NotNil(t, ns)
		assert.Equal(t, "finance", ns.Key)
	})

	t.Run("Example_ListNamespacesInFamily", func(t *testing.T) {
		// SCENARIO: The planner creates multiple namespaces and reviews them all
		doRequest(t, router, "PUT", "/family/bigcorp", map[string]interface{}{"valid_from": "now"})
		doRequest(t, router, "PUT", "/namespace/bigcorp/finance", map[string]interface{}{"valid_from": "now"})
		doRequest(t, router, "PUT", "/namespace/bigcorp/operations", map[string]interface{}{"valid_from": "now"})
		doRequest(t, router, "PUT", "/namespace/bigcorp/marketing", map[string]interface{}{"valid_from": "now"})

		// List all namespaces in bigcorp
		resp, body := doRequest(t, router, "GET", "/namespaces/bigcorp", nil)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result namespacesOutput
		require.NoError(t, json.Unmarshal(body, &result))
		assert.GreaterOrEqual(t, len(result.Namespaces), 3)
	})

	t.Run("Example_NamespaceMustHaveValidParent", func(t *testing.T) {
		// SCENARIO: The planner tries to create a namespace in a family that doesn't exist
		resp, body := doRequest(t, router, "PUT", "/namespace/nonexistentfam/finance", map[string]interface{}{
			"valid_from": "now",
		})

		// The API rejects it because the parent family doesn't exist
		assert.Equal(t, http.StatusNotFound, resp.StatusCode, "response: %s", string(body))
	})

	t.Run("Example_RetrieveSpecificNamespace", func(t *testing.T) {
		// SCENARIO: The planner wants to check details of a specific namespace
		doRequest(t, router, "PUT", "/family/startup", map[string]interface{}{"valid_from": "now"})
		doRequest(t, router, "PUT", "/namespace/startup/engineering", map[string]interface{}{"valid_from": "now"})

		// Get the specific namespace
		resp, body := doRequest(t, router, "GET", "/namespace/startup/engineering", nil)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result namespaceOutput
		require.NoError(t, json.Unmarshal(body, &result))
		assert.Equal(t, "engineering", result.Key)
	})
}
