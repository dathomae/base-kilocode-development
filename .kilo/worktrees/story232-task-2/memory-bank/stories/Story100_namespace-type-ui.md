# Story100: Namespace Type Selection in UI

## Goal

Add namespace type selection to the UI when creating a new namespace, and display the namespace type in the hierarchy view. This will allow users to choose from predefined namespace types (Production, Smoke Test, Integration, UAT, Alpha, Scenario, Experiment, Beta) instead of accepting the default "Production" type.

## Background

The `namespace` table has an `ns_type_id` field that references the `namespace_type` table. The `namespace_type` table contains eight predefined values representing different environment types:

| ns_type_id | Name        |
| ---------- | ----------- |
| 1          | Production  |
| 2          | Smoke Test  |
| 3          | Integration |
| 4          | UAT         |
| 5          | Alpha       |
| 6          | Scenario    |
| 7          | Experiment  |
| 8          | Beta        |

Currently, the backend defaults to `ns_type_id=1` (Production) when creating a namespace. The UI does not provide a way for users to select a different namespace type, nor does it display the namespace type in the hierarchy view.

## References

- [Policy Schema](../../src/backend/serde/rdbms/schema.sql) - Defines `namespace_type` and `namespace` tables
- [Metrics Service](../../src/backend/src/metrics/metrics_service.py) - Contains default namespace type logic
- [Frontend API Service](../../src/frontend/src/services/metricsApi.ts) - API client for metadata operations
- [MetadataTree Component](../../src/frontend/src/components/MetadataTree.tsx) - Tree view component
- [Definitions Page](../../src/frontend/src/pages/Definitions.tsx) - Main definitions page

## Dependencies

- [Story009_frontend-service-integration.md](Story009_frontend-service-integration.md) - Completed (requires frontend with Metrics Service integration)

## Dependent Stories

(To be populated as other stories are added)

## Tasks

1. Add backend endpoint for namespace types
   a. Create a new GET `/namespace-types` endpoint in the Metrics Service
   b. Return the list of namespace types with their IDs and names
   c. Add appropriate tests for the new endpoint

2. Update frontend API service
   a. Add `NamespaceType` interface to [`metricsApi.ts`](../../src/frontend/src/services/metricsApi.ts)
   b. Add `fetchNamespaceTypes()` function to retrieve namespace types
   c. Add `ns_type_name` field to [`HierarchyItem`](../../src/frontend/src/services/metricsApi.ts:51) interface

3. Update backend hierarchy endpoint
   a. Modify the `/list/hierarchy` endpoint to include namespace type name in the response
   b. Join with `namespace_type` table to get the type name
   c. Update tests for the modified endpoint

4. Update MetadataTree component for namespace type display
   a. Display namespace type as a badge or label next to namespace names in the tree
   b. Use distinct styling for different namespace types (optional enhancement)
   c. Consider color coding for quick visual identification (e.g., Production = green, Test types = blue)

5. Add namespace type selector to the add dialog
   a. Modify the add child dialog in [`MetadataTree.tsx`](../../src/frontend/src/components/MetadataTree.tsx) to show a dropdown when adding a namespace
   b. Fetch namespace types from the new endpoint
   c. Default to "Production" but allow user selection
   d. Pass the selected namespace type to the create metadata function

6. Update Definitions page for namespace creation
   a. Modify [`Definitions.tsx`](../../src/frontend/src/pages/Definitions.tsx) to handle namespace type in the `handleAddChild` function
   b. Pass `ns_type_id` as an attribute when creating a namespace
   c. Update the attributes handling to support namespace type selection

7. Update backend to accept namespace type from frontend
   a. Modify the metadata creation logic in [`metrics_service.py`](../../src/backend/src/metrics/metrics_service.py) to accept `ns_type_id` from the request
   b. Validate that the provided `ns_type_id` exists in the `namespace_type` table
   c. Fall back to default (Production) if not provided, for backward compatibility
   d. Add tests for namespace creation with explicit type

8. Add frontend tests
   a. Add tests for `fetchNamespaceTypes()` function
   b. Add tests for namespace type selector component
   c. Update existing tests to account for namespace type in hierarchy items

## Notes

### Namespace Type Display Considerations

The namespace type should be displayed in a way that is informative but not cluttered. Options include:

1. **Badge style**: A small colored badge next to the namespace name (e.g., "MyNamespace [Production]")
2. **Icon style**: Different icons for different types (e.g., star for Production, flask for Experiment)
3. **Tooltip**: Show the type in a tooltip on hover

The recommended approach is a badge style with optional color coding for quick visual scanning.

### Default Behavior

For backward compatibility and ease of use, the namespace type selector should default to "Production" (ns_type_id=1). This matches the current backend behavior and is the most common use case.

### Validation

The backend should validate that the provided `ns_type_id` is a valid value from the `namespace_type` table. If an invalid ID is provided, the API should return a 400 Bad Request error with a clear message.

### Future Enhancements

- Allow filtering the hierarchy by namespace type
- Add namespace type as a column in table views (if implemented)
- Support bulk namespace type changes (update functionality)
