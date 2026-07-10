package handlers

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Example: A planner controls the time validity of planning entities
func TestTimestampParsing(t *testing.T) {
	_, router := createTestAPI(t, testDB)

	t.Run("Example_EntityValidFromNow", func(t *testing.T) {
		// SCENARIO: The planner creates an entity effective immediately ("now")
		// This is useful for entities being created in real time
		resp, body := doRequest(t, router, "PUT", "/family/immediate", map[string]interface{}{
			"valid_from": "now",
		})

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result familyOutput
		require.NoError(t, json.Unmarshal(body, &result))

		// The entity is valid from approximately the current time
		assert.True(t, time.Since(result.ValidityStart) < time.Minute, "validity_start should be recent")
	})

	t.Run("Example_EntityValidFromFutureDate", func(t *testing.T) {
		// SCENARIO: The planner creates an entity effective at a specific future date
		// This is useful for retroactive or planned planning changes
		futureDate := "2025-06-01T12:00:00Z"
		resp, body := doRequest(t, router, "PUT", "/family/futureplan", map[string]interface{}{
			"valid_from": futureDate,
		})

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result familyOutput
		require.NoError(t, json.Unmarshal(body, &result))

		// The entity is valid from the specified date
		assert.Equal(t, 2025, result.ValidityStart.Year())
		assert.Equal(t, 6, int(result.ValidityStart.Month()))
		assert.Equal(t, 1, result.ValidityStart.Day())
	})

	t.Run("Example_EntityValidityWindow", func(t *testing.T) {
		// SCENARIO: The planner creates an entity that is valid for a time window
		// E.g., "This plan is valid from now through the end of 2024"
		resp, body := doRequest(t, router, "PUT", "/family/temporary", map[string]interface{}{
			"valid_from": "now",
			"valid_to":   "infinity", // Using infinity as "far in the future"
		})

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result familyOutput
		require.NoError(t, json.Unmarshal(body, &result))

		// The entity is valid from now to infinity
		assert.True(t, time.Since(result.ValidityStart) < time.Minute)
		// infinity translates to 3000-01-01 internally
		assert.Equal(t, 3000, result.ValidityEnd.Year())
	})

	t.Run("Example_InvalidTimestampFormatRejected", func(t *testing.T) {
		// SCENARIO: The planner provides a timestamp in an unsupported format
		resp, body := doRequest(t, router, "PUT", "/family/badtime", map[string]interface{}{
			"valid_from": "sometime-in-the-future", // Not a valid format
		})

		// The API rejects the request with validation error
		assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode, "response: %s", string(body))
	})
}
