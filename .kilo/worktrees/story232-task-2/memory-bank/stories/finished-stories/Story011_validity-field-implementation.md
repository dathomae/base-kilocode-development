# Story011: Validity Field Implementation

## Goal

Re-enable the validity fields in the SQL schema files where they have been commented out, and implement the corresponding logic in the Metrics Service to automatically set validity ranges when entities are created or updated. The validity range will be set with the lower bound to the current transaction timestamp and the upper bound to infinity, representing "valid from now until further notice."

## Background

The validity field is a `tstzrange` (time range with time zone) column that implements a bitemporal model for tracking when entity versions are valid. Several tables had their validity fields commented out during initial development (see [open_questions.md](../../resources/system_concepts/open_questions.md)). The decision has been made to keep these fields in the schema but implement a simplified approach where:

- **Lower bound**: Set to the current transaction timestamp at creation time
- **Upper bound**: Set to infinity (representing "valid indefinitely until explicitly closed")

This approach allows for future implementation of full bitemporal versioning without requiring immediate complexity.

## References

- [Architecture](../specifications/architecture/)
- [Requirements](../specifications/requirements/)
- [Open Questions - Validity Field Discussion](../../resources/system_concepts/open_questions.md)
- [Taxonomy Schema](../../src/backend/schema/taxonomy.sql)
- [Policy Schema](../../src/backend/schema/policy.sql)

## Dependencies

- [Story006_metrics-service-database-integration.md](Story006_metrics-service-database-integration.md) - Completed (requires Metrics Service with database integration)
- [Story010_metadata-update-delete.md](Story010_metadata-update-delete.md) - Not Started (may need coordination if update logic changes)

## Dependent Stories

(To be populated as other stories are added)

## Tasks

1. Uncomment validity fields in policy.sql - Completed
   a. Uncomment `validity` column definition in `namespace_family` table - Completed
   b. Uncomment `validity` column definition in `namespace` table - Completed
   c. Uncomment the EXCLUDE constraints for both tables that use validity - Completed
   d. Comment out or remove the simplified PRIMARY KEY and UNIQUE constraints that replaced the EXCLUDE constraints - Completed
   e. Uncomment the accelerator indexes that reference `UPPER(validity)` - Completed
   f. Verify the SQL syntax is correct after uncommenting - Completed

2. Uncomment validity fields in taxonomy.sql - Completed
   a. Uncomment `validity` column definition in `category` table - Completed
   b. Uncomment the EXCLUDE constraints for `category` that use validity - Completed
   c. Comment out or remove the simplified PRIMARY KEY and UNIQUE constraints that replaced the EXCLUDE constraints - Completed
   d. Uncomment the accelerator indexes that reference `UPPER(validity)` - Completed
   e. Verify the SQL syntax is correct after uncommenting - Completed

3. Update Metrics Service models for validity - Completed
   a. Add `validity` field to Pydantic models for `NamespaceFamily`, `Namespace`, and `Category` in `models.py` - Completed
   b. Make the validity field optional in request models (will be auto-generated) - Completed
   c. Include validity in response models for completeness - Completed
   d. Add tests for the updated models - Completed

4. Update repository layer for validity handling - Completed
   a. Modify `create_namespace_family()` to set validity to `[current_timestamp, infinity)` - Completed
   b. Modify `create_namespace()` to set validity to `[current_timestamp, infinity)` - Completed
   c. Modify `create_category()` to set validity to `[current_timestamp, infinity)` - Completed
   d. Use PostgreSQL's `tstzrange(current_timestamp, NULL, '[)')` or equivalent for infinity - Completed
   e. Add tests for the validity range generation in repository tests - Completed

5. Update existing tests for schema changes - Completed
   a. Run existing test suite to identify failures after schema changes - Completed
   b. Update test fixtures to accommodate validity field - Completed
   c. Update any hardcoded SQL queries in tests - Completed
   d. Ensure all tests pass with the new schema - Completed

6. Update open_questions.md - Completed
   a. Mark the validity field question as resolved - Completed
   b. Document the decision to use `[current_timestamp, infinity)` as the default validity range - Completed
   c. Note that full bitemporal versioning can be implemented in the future if needed - Completed

7. Reinitialize the database - Completed
   a. Drop existing tables (no production data to preserve) - Completed
   b. Run database initialization CLI to recreate tables with updated schema - Completed
   c. Verify all tables are created correctly with validity columns - Completed
   d. Verify all constraints and indexes are created correctly - Completed
   e. Test entity creation through the API and verify validity is set correctly - Completed

8. Uncomment validity fields in morphisms.sql - Completed
   a. Uncomment `validity` column definition in `metric` table - Completed
   b. Uncomment `validity` column definition in `metric_dim` table - Completed
   c. Uncomment `validity` column definition in `transformer` table - Completed
   d. Uncomment `validity` column definition in `transformer_input` table - Completed
   e. Uncomment `validity` column definition in `transformer_output` table - Completed
   f. Uncomment the EXCLUDE constraints for all five tables that use validity - Completed
   g. Comment out or remove the simplified PRIMARY KEY and UNIQUE constraints that replaced the EXCLUDE constraints - Completed
   h. Uncomment the accelerator indexes that reference `UPPER(validity)` - Completed
   i. Verify the SQL syntax is correct after uncommenting - Completed

9. (Merged into morphisms.sql) - Completed
   a. Uncomment `validity` column definition in `shifter` table - Completed
   b. Uncomment `validity` column definition in `shifter_input` table - Completed
   c. Uncomment `validity` column definition in `shifter_output` table - Completed
   d. Uncomment the EXCLUDE constraints for all three tables that use validity - Completed
   e. Comment out or remove the simplified PRIMARY KEY and UNIQUE constraints that replaced the EXCLUDE constraints - Completed
   f. Uncomment the accelerator indexes that reference `UPPER(validity)` - Completed
   g. Verify the SQL syntax is correct after uncommenting - Completed

10. Update Metrics Service models for transform/shift entities - Completed
    a. Add `validity` field to Pydantic models for `Metric`, `MetricDim`, `Transformer`, `TransformerInput`, `TransformerOutput` - Completed
    b. Add `validity` field to Pydantic models for `Shifter`, `ShifterInput`, `ShifterOutput` - Completed
    c. Make the validity field optional in request models (will be auto-generated) - Completed
    d. Include validity in response models for completeness - Completed
    e. Add appropriate validators for SMALLINT fields - Completed

11. Update repository layer for transform/shift validity handling - Completed
    a. Added table mappings for all new entity types - Completed
    b. Added column mappings for all new entity types - Completed
    c. Updated \_ID_COLUMNS to support composite primary keys (returns None for composite keys) - Completed
    d. Updated \_get_id_column to return Optional[str] - Completed
    e. Updated create_metadata to handle composite key tables (returns None for them) - Completed
    f. Updated update_metadata to raise ValueError for composite key tables - Completed
    g. Updated delete_metadata to raise ValueError for composite key tables - Completed
    h. Updated get_metadata_by_id to raise ValueError for composite key tables - Completed

## Notes

### Affected Tables

The following tables have validity fields currently commented out that need to be re-enabled:

| Table                | Schema File     | Entity Type        | Status    |
| -------------------- | --------------- | ------------------ | --------- |
| `namespace_family`   | `policy.sql`    | Namespace Family   | Completed |
| `namespace`          | `policy.sql`    | Namespace          | Completed |
| `category`           | `taxonomy.sql`  | Category           | Completed |
| `metric`             | `morphisms.sql` | Metric             | Completed |
| `metric_dim`         | `morphisms.sql` | Metric Dimension   | Completed |
| `transformer`        | `morphisms.sql` | Transformer        | Completed |
| `transformer_input`  | `morphisms.sql` | Transformer Input  | Completed |
| `transformer_output` | `morphisms.sql` | Transformer Output | Completed |
| `shifter`            | `morphisms.sql` | Shifter            | Completed |
| `shifter_input`      | `morphisms.sql` | Shifter Input      | Completed |
| `shifter_output`     | `morphisms.sql` | Shifter Output     | Completed |

### Tables with Validity Already Enabled

The following tables already have validity fields in place and are not affected by this story:

- `ts_doc_type`, `ts_doc_row_type` (timeseries.sql)
- `category_attr`, `element`, `element_attr` (taxonomy.sql)
- `storage_config` (connectors.sql)

### PostgreSQL Validity Range Syntax

To create a validity range from current time to infinity:

```sql
-- Using tstzrange with NULL for infinity
tstzrange(transaction_timestamp(), NULL, '[)')

-- Or using the range literal syntax
'[current_timestamp,)'::tstzrange  -- Note: infinity is represented as unbounded
```

### Database Reinitialization

Since there is no production data to preserve, the database will be dropped and reinitialized after the SQL schema changes are made. This simplifies the implementation by avoiding the need for migration scripts.

### Future Considerations

- The current implementation sets validity to infinity, which means entities are "valid forever" until explicitly closed
- Full bitemporal versioning (close-and-insert pattern for updates) can be implemented later
- The `get_taxonomy_tree()` function in taxonomy.sql already uses validity checks in its temporal queries
