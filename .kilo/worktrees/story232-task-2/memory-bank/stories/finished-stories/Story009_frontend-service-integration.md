# Story009: Frontend Service Integration

## Goal

Modify the frontend to read from and write to the Metrics Service endpoints instead of using hardcoded in-memory data. The Definitions page will fetch metadata types from the `/types` endpoint, retrieve existing metadata from the `/list` endpoint, and persist new metadata entries to the `/metadata` endpoint. This replaces the current in-memory storage with actual backend persistence.

## References

- [Architecture](../specifications/architecture/)
- [Requirements](../specifications/requirements/)
- [Story003_frontend-structure.md](Story003_frontend-structure.md)
- [Story004_core-and-metrics-services.md](Story004_core-and-metrics-services.md)
- [Story006_metrics-service-database-integration.md](Story006_metrics-service-database-integration.md)
- [Story007_ui-navigation-and-pages.md](Story007_ui-navigation-and-pages.md)

## Dependencies

- [Story003_frontend-structure.md](Story003_frontend-structure.md) - Completed (requires React project setup with Vite and Vitest)
- [Story004_core-and-metrics-services.md](Story004_core-and-metrics-services.md) - Completed (requires Metrics Service with `/list`, `/types`, and `/metadata` endpoints)
- [Story006_metrics-service-database-integration.md](Story006_metrics-service-database-integration.md) - Completed (requires Metrics Service to actually persist data to database)
- [Story007_ui-navigation-and-pages.md](Story007_ui-navigation-and-pages.md) - Completed (requires Definitions page with MetadataTable component)

## Dependent Stories

- [Story012_openapi-compliance.md](Story012_openapi-compliance.md) - Not Started

## Tasks

1. Remove dummy data from Definitions page - Completed
   a. Remove the `initialMetadataTypes` array from `src/pages/Definitions.tsx` - Completed
   b. Update the intro text to remove the statement about local storage (line 113-114) - Completed
   c. Start with empty metadata types array until API data is loaded - Completed
   d. Verify no hardcoded items remain in the component - Completed

2. Create API service module for Metrics Service - Completed
   a. Create `src/services/metricsApi.ts` for Metrics Service API calls - Completed
   b. Define TypeScript interfaces for API request/response types - Completed
   c. Implement `fetchMetadataTypes()` function to call `/types` endpoint - Completed
   d. Implement `fetchMetadataList(type?)` function to call `/list` endpoint - Completed
   e. Implement `createMetadata(metadataType, name, description)` function to call `/metadata` endpoint - Completed
   f. Add error handling with appropriate error types - Completed
   g. Add logging for API calls in development mode - Completed

3. Configure environment variables for Metrics Service - Completed
   a. Add `VITE_METRICS_SERVICE_URL` to `.env.example` file (default: `http://localhost:8001`) - Completed
   b. Update `.env` file with actual Metrics Service URL for local development - Completed
   c. Update `src/config.ts` to export Metrics Service URL - Completed
   d. Add `getMetricsApiUrl(path)` helper function to config - Completed
   e. Verify environment variables are properly typed - Completed

4. Update Definitions page to fetch metadata types - Completed
   a. Modify `src/pages/Definitions.tsx` to call `fetchMetadataTypes()` on component mount - Completed
   b. Add loading state for initial metadata type fetch - Completed
   c. Add error state with user-friendly error message if fetch fails - Completed
   d. Transform API response to match `MetadataType` interface expected by `MetadataTable` - Completed
   e. Test that metadata types load correctly from the API - Completed

5. Update Definitions page to fetch existing metadata - Completed
   a. Modify `src/pages/Definitions.tsx` to call `fetchMetadataList()` for each metadata type - Completed
   b. Add loading state for metadata items fetch - Completed
   c. Merge fetched items with their corresponding metadata types - Completed
   d. Handle empty responses gracefully (show empty state) - Completed
   e. Test that existing metadata displays correctly in tables - Completed

6. Update Definitions page to persist new metadata - Completed
   a. Modify `handleAddItem` callback to call `createMetadata()` API instead of local state update - Completed
   b. Add loading state during save operation - Completed
   c. On successful save, refresh the metadata list to show the new item - Completed
   d. On save failure, display error message and keep edit form open - Completed
   e. Test that new metadata is persisted and appears after page refresh - Completed

7. Update MetadataTable component - Completed
   a. Add loading state prop to disable add button during save operations - Completed
   b. Update tests to verify loading state behavior - Completed
   c. Verify component handles empty items array correctly - Completed

8. Add API error handling and user feedback - Completed
   a. Create `src/components/ErrorMessage.tsx` component for displaying API errors - Completed
   b. Create `src/components/LoadingSpinner.tsx` component for loading states - Completed
   c. Integrate error display into Definitions page - Completed
   d. Add retry functionality for failed fetches - Completed
   e. Test error states render correctly - Completed

9. Update tests for service integration - Completed
   a. Add MSW (Mock Service Worker) or similar for mocking API calls in tests - Completed
   b. Update `Definitions.test.tsx` to mock `metricsApi` service - Completed
   c. Add tests for successful data fetch scenarios - Completed
   d. Add tests for API error scenarios - Completed
   d. Add tests for metadata creation flow - Completed
   f. Create `metricsApi.test.ts` for unit testing API service functions - Completed
   g. Run all tests and verify they pass - Completed

10. Update documentation - Completed
    a. Update frontend README with Metrics Service integration details - Completed
    b. Document the API service module structure - Completed
    c. Add troubleshooting section for common API connection issues - Completed
    d. Document environment variables required for Metrics Service connection - Completed

11. Verify complete integration - Completed
    a. Start Metrics Service on port 8001 - Completed
    b. Start frontend development server - Completed
    c. Verify Definitions page loads metadata types from `/types` endpoint - Completed
    d. Verify Definitions page loads metadata from `/list` endpoint - Completed
    e. Add a new metadata entry and verify it persists via `/metadata` endpoint - Completed
    f. Refresh the page and verify the new entry still appears - Completed
    g. Run full test suite to ensure all tests pass - Completed
    h. Build production bundle to verify no build errors - Completed

## Notes

- The current Definitions page stores all data in React state with hardcoded initial values in `initialMetadataTypes` array
- The `MetadataTable` component currently receives items through props and calls `onAddItem` callback with new item data
- Metrics Service endpoints:
  - `POST /types` - Returns list of supported metadata types
  - `POST /list` - Returns metadata items, optionally filtered by type
  - `POST /metadata` - Creates new metadata entry
- The frontend config module (`src/config.ts`) already has pattern for API URL construction via `getApiUrl()`
- Environment variables must be prefixed with `VITE_` to be exposed to client code in Vite
- The transformation from API response format to UI format may require mapping functions
- Loading states should be shown during API calls to provide user feedback
- Error handling should distinguish between network errors and API errors
- MSW (Mock Service Worker) is recommended for mocking API calls in tests as it intercepts requests at the network level
- The existing `MetadataItem` and `MetadataType` interfaces may need adjustment to match API response structure
