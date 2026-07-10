# Story010: Metadata Update and Delete Functions

## Goal

Add update and delete functionality for metadata entities to the Metrics Service API and the Definitions page UI. This includes creating new API endpoints for updating and deleting metadata, adding corresponding functions to the frontend API service, and adding edit and delete icons to each row in the metadata tables. Since the frontend does not receive internal database IDs, the update endpoint must accept both old and new names along with entity types to correctly identify records.

## References

- [Architecture](../specifications/architecture/)
- [Requirements](../specifications/requirements/)
- [Story004_core-and-metrics-services.md](Story004_core-and-metrics-services.md)
- [Story007_ui-navigation-and-pages.md](Story007_ui-navigation-and-pages.md)
- [Story009_frontend-service-integration.md](Story009_frontend-service-integration.md)

## Dependencies

- [Story004_core-and-metrics-services.md](Story004_core-and-metrics-services.md) - Completed (requires Metrics Service with `/metadata` endpoint)
- [Story007_ui-navigation-and-pages.md](Story007_ui-navigation-and-pages.md) - Completed (requires Definitions page with MetadataTable component)
- [Story009_frontend-service-integration.md](Story009_frontend-service-integration.md) - Completed (requires frontend API service and backend integration)

## Dependent Stories

(To be populated as other stories are added)

## Tasks

1. Add update endpoint to Metrics Service - Not Started
   a. Create `PUT /metadata` endpoint in `metrics_service.py` - Not Started
   b. Define `MetadataUpdateRequest` Pydantic model with `old_name`, `new_name`, `metadata_type`, `description`, and optional `attributes` fields - Not Started
   c. Implement name-based record lookup (since IDs are not sent to frontend) - Not Started
   d. Call `update_metadata` from repository with resolved record ID - Not Started
   e. Add error handling for not-found and conflict scenarios - Not Started
   f. Add tests for the update endpoint in `test_metadata.py` - Not Started

2. Add delete endpoint to Metrics Service - Not Started
   a. Create `DELETE /metadata` endpoint in `metrics_service.py` - Not Started
   b. Define `MetadataDeleteRequest` Pydantic model with `name` and `metadata_type` fields - Not Started
   c. Implement name-based record lookup to resolve record ID - Not Started
   d. Call `delete_metadata` from repository with resolved record ID - Not Started
   e. Add error handling for not-found scenarios - Not Started
   f. Add tests for the delete endpoint in `test_metadata.py` - Not Started

3. Add update function to frontend API service - Not Started
   a. Create `updateMetadata(oldName, newName, metadataType, description, attributes?)` function in `metricsApi.ts` - Not Started
   b. Define `MetadataUpdateRequest` and `MetadataUpdateResponse` TypeScript interfaces - Not Started
   c. Implement PUT request to `/metadata` endpoint - Not Started
   d. Add error handling for update failures - Not Started
   e. Add tests for `updateMetadata` in `metricsApi.test.ts` - Not Started

4. Add delete function to frontend API service - Not Started
   a. Create `deleteMetadata(name, metadataType)` function in `metricsApi.ts` - Not Started
   b. Define `MetadataDeleteRequest` and `MetadataDeleteResponse` TypeScript interfaces - Not Started
   c. Implement DELETE request to `/metadata` endpoint with request body - Not Started
   d. Add error handling for delete failures - Not Started
   e. Add tests for `deleteMetadata` in `metricsApi.test.ts` - Not Started

5. Add edit functionality to MetadataTable component - Not Started
   a. Add edit icon button to each metadata row (next to the name/description) - Not Started
   b. Add `onEditItem` callback prop to `MetadataTableProps` interface - Not Started
   c. Implement inline editing mode when edit icon is clicked (similar to add mode) - Not Started
   d. Pre-populate edit form with existing name and description values - Not Started
   e. Call `onEditItem` callback on save with old and new values - Not Started
   f. Add loading state during edit operations - Not Started
   g. Update tests in `MetadataTable.test.tsx` for edit functionality - Not Started

6. Add delete functionality to MetadataTable component - Not Started
   a. Add delete icon button to each metadata row (next to edit icon) - Not Started
   b. Add `onDeleteItem` callback prop to `MetadataTableProps` interface - Not Started
   c. Implement confirmation dialog before delete (browser native `confirm()` is acceptable) - Not Started
   d. Call `onDeleteItem` callback on confirmation - Not Started
   e. Add loading state during delete operations - Not Started
   f. Update tests in `MetadataTable.test.tsx` for delete functionality - Not Started

7. Update Definitions page to handle edit and delete - Not Started
   a. Implement `handleEditItem` callback in `Definitions.tsx` - Not Started
   b. Call `updateMetadata` API and refresh metadata list on success - Not Started
   c. Show error message and keep edit form open on failure - Not Started
   d. Implement `handleDeleteItem` callback in `Definitions.tsx` - Not Started
   e. Call `deleteMetadata` API and refresh metadata list on success - Not Started
   f. Show error message on delete failure - Not Started
   g. Update tests in `Definitions.test.tsx` for edit and delete flows - Not Started

8. Add CSS styling for edit and delete icons - Not Started
   a. Add styles for edit icon button in `MetadataTable.css` - Not Started
   b. Add styles for delete icon button in `MetadataTable.css` - Not Started
   c. Ensure icons are accessible with proper hover states and focus indicators - Not Started
   d. Ensure icons do not clutter the table layout on smaller screens - Not Started

9. Verify complete integration - Not Started
   a. Start Metrics Service on port 8001 - Not Started
   b. Start frontend development server - Not Started
   c. Verify edit functionality works end-to-end - Not Started
   d. Verify delete functionality works end-to-end - Not Started
   e. Verify error handling works correctly for both operations - Not Started
   f. Run full test suite to ensure all tests pass - Not Started
   g. Build production bundle to verify no build errors - Not Started

## Notes

- The backend repository already has `update_metadata()` and `delete_metadata()` functions in `repository.py` that work with record IDs
- Since the frontend does not receive internal database IDs, the API must look up records by name and type
- The update endpoint needs both `old_name` (to find the record) and `new_name` (the updated value)
- For hierarchical metadata types (category, element, metric), the name lookup may need additional context (e.g., namespace name, category name) to uniquely identify records
- The `MetadataTable` component currently displays items in a read-only table with only add functionality
- Edit and delete icons should be subtle but accessible - consider using SVG icons or Unicode characters (✎ for edit, 🗑 for delete)
- Loading states should disable the relevant row's action buttons during operations
- Consider adding optimistic updates for better UX, but ensure error handling reverts the UI on failure
- The existing API service pattern in `metricsApi.ts` should be followed for consistency
