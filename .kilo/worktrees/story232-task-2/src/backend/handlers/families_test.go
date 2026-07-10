package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"chronocone/backend/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Example: A planner creates the root family for ACME Corp's planning system
func TestFamilyHandlers(t *testing.T) {
	_, router := createTestAPI(t, testDB)
	ctx := context.Background()

	t.Run("Example_CreateAndRetrieveAFamily", func(t *testing.T) {
		// SCENARIO: A planner sets up a new planning hierarchy root
		// BEFORE: No "acmecorp" family in the system

		// The planner creates the family
		createResp, createBody := doRequest(t, router, "PUT", "/family/acmecorp", map[string]interface{}{
			"valid_from": "now",
		})

		// AFTER: The family is created and returned
		assert.Equal(t, http.StatusOK, createResp.StatusCode)
		var created familyOutput
		require.NoError(t, json.Unmarshal(createBody, &created))
		assert.Equal(t, "acmecorp", created.Key)

		// The planner can retrieve it by key
		getResp, getBody := doRequest(t, router, "GET", "/family/acmecorp", nil)
		assert.Equal(t, http.StatusOK, getResp.StatusCode)

		var retrieved familyOutput
		require.NoError(t, json.Unmarshal(getBody, &retrieved))
		assert.Equal(t, "acmecorp", retrieved.Key)

		// Verify it persisted to the database via the domain layer
		family, err := domain.GetFamilyByKey(ctx, testDB, "acmecorp", nil, nil)
		require.NoError(t, err)
		require.NotNil(t, family)
		assert.Equal(t, "acmecorp", family.Key)
	})

	t.Run("Example_ListMultipleFamilies", func(t *testing.T) {
		// SCENARIO: The planner has created multiple families and wants to see all of them
		doRequest(t, router, "PUT", "/family/acme", map[string]interface{}{"valid_from": "now"})
		doRequest(t, router, "PUT", "/family/widgets", map[string]interface{}{"valid_from": "now"})
		doRequest(t, router, "PUT", "/family/globex", map[string]interface{}{"valid_from": "now"})

		// List all families
		resp, body := doRequest(t, router, "GET", "/families", nil)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result familiesOutput
		require.NoError(t, json.Unmarshal(body, &result))
		assert.GreaterOrEqual(t, len(result.Families), 3)
	})

	t.Run("Example_FamilyKeyMustBeValid", func(t *testing.T) {
		// SCENARIO: The planner tries to create a family with an invalid key

		// Keys that are too short (< 6 chars) are rejected
		resp, body := doRequest(t, router, "PUT", "/family/short", map[string]interface{}{
			"valid_from": "now",
		})
		assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode, "short keys rejected: %s", string(body))

		// Keys that are too long (> 64 chars) are rejected
		longKey := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" // 65 chars
		resp, body = doRequest(t, router, "PUT", fmt.Sprintf("/family/%s", longKey), map[string]interface{}{
			"valid_from": "now",
		})
		assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode, "long keys rejected: %s", string(body))
	})

	t.Run("Example_FamilyNotFound", func(t *testing.T) {
		// SCENARIO: The planner tries to retrieve a family that doesn't exist
		resp, body := doRequest(t, router, "GET", "/family/nonexistent", nil)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode, "response: %s", string(body))
	})

	t.Run("Example_UpdateFamilyValidity", func(t *testing.T) {
		// SCENARIO: A planner marks a family as ending at a future date
		// BEFORE: Family exists with indefinite validity

		doRequest(t, router, "PUT", "/family/deprecated", map[string]interface{}{
			"valid_from": "now",
		})

		// The planner updates it to end at a specific time
		patchResp, patchBody := doRequest(t, router, "PATCH", "/family/deprecated", map[string]interface{}{
			"valid_to": "infinity", // marks as ending far in future
		})

		// AFTER: The update is applied
		assert.Equal(t, http.StatusOK, patchResp.StatusCode)

		var result familyOutput
		require.NoError(t, json.Unmarshal(patchBody, &result))
		assert.Equal(t, "deprecated", result.Key)
		// validity_end should be set to infinity (3000-01-01)
		assert.True(t, result.ValidityEnd.Year() >= 3000)
	})
}
