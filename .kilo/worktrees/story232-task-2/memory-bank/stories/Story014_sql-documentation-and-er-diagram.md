# Story014: SQL Schema Documentation and ER Diagram

## Goal

Add comprehensive comments to all SQL schema sections to explain the purpose of each attribute and the relationships to other tables. Additionally, generate an Entity-Relationship (ER) diagram to be included in the architecture documentation, providing a visual representation of the database schema.

## Background

The current SQL schema files (now consolidated into a single master schema) contain minimal documentation. While some tables have "APPLICATION CONSTRAINT" comments noting relationships that cannot be expressed as foreign keys (due to bitemporal validity ranges), most columns lack explanations of their purpose. This makes it difficult for developers and AI assistants to understand the data model without extensive cross-referencing.

An ER diagram will provide a high-level visual overview of the database structure, making it easier to understand table relationships at a glance. This is particularly valuable for onboarding new team members and for reference during development.

## References

- [Architecture](../specifications/architecture/)
- [Consolidated Schema](../../src/backend/serde/rdbms/schema.sql)

## Dependencies

- [Story005_database-initialization-cli.md](Story005_database-initialization-cli.md) - Completed (requires stable schema structure)
- [Story011_validity-field-implementation.md](Story011_validity-field-implementation.md) - Not Started (may affect validity field comments)

## Dependent Stories

(To be populated as other stories are added)

## Tasks

1. Document Policy section (in schema.sql) - Not Started
   a. Add table-level comment block explaining the policy domain (namespaces, subjects, access control) - Not Started
   b. Document `namespace_family_type` table columns and their purpose - Not Started
   c. Document `namespace_type` table columns and their purpose - Not Started
   d. Document `subject_gateway` table columns and their purpose - Not Started
   e. Document `namespace_family_namespace_type` table columns and their purpose - Not Started
   f. Document `namespace_family` table columns and their purpose - Not Started
   g. Document `namespace` table columns and their purpose - Not Started
   h. Add relationship comments explaining foreign key connections - Not Started
   i. Verify SQL syntax is correct after adding comments - Not Started

2. Document Taxonomy section (in schema.sql) - Not Started
   a. Add table-level comment block explaining the taxonomy domain (categories, elements, attributes) - Not Started
   b. Document `category` table columns and their purpose - Not Started
   c. Document `category_attr` table columns and their purpose - Not Started
   d. Document `element` table columns and their purpose - Not Started
   e. Document `element_attr` table columns and their purpose - Not Started
   f. Document `get_taxonomy_tree` function parameters and return values - Not Started
   g. Add relationship comments explaining foreign key connections - Not Started
   h. Verify SQL syntax is correct after adding comments - Not Started

3. Document Timeseries section (in schema.sql) - Not Started
   a. Add table-level comment block explaining the timeseries domain (documents, row types, lineage) - Not Started
   b. Document `ts_doc_type` table columns and their purpose - Not Started
   c. Document `ts_doc_row_type` table columns and their purpose - Not Started
   d. Document `ts_doc` table columns and their purpose - Not Started
   e. Document `ts_doc_tags` table columns and their purpose - Not Started
   f. Document `ts_doc_input_lineage` table columns and their purpose - Not Started
   g. Add relationship comments explaining foreign key connections - Not Started
   h. Verify SQL syntax is correct after adding comments - Not Started

4. Document Connectors section (in schema.sql) - Not Started
   a. Add table-level comment block explaining the connectors domain (storage configuration) - Not Started
   b. Document `storage_config` table columns and their purpose - Not Started
   c. Add relationship comments explaining how storage_config connects to other tables - Not Started
   d. Verify SQL syntax is correct after adding comments - Not Started

5. Document Morphisms section (in schema.sql) - Not Started
   a. Add table-level comment block explaining the morphisms domain (metrics, transforms, data flow) - Not Started
   b. Document `metric` table columns and their purpose - Not Started
   c. Document `metric_dim` table columns and their purpose - Not Started
   d. Document `transform` table columns and their purpose - Not Started
   e. Document `transform_input` table columns and their purpose - Not Started
   f. Document `transform_output` table columns and their purpose - Not Started
   g. Add relationship comments explaining foreign key connections - Not Started
   h. Verify SQL syntax is correct after adding comments - Not Started

6. Generate ER diagram - Not Started
   a. Evaluate ER diagram tools (e.g., D2, dbdiagram.io, Mermaid, DrawSQL) - Not Started
   b. Create ER diagram source file in appropriate format - Not Started
   c. Include all tables from the master schema - Not Started
   d. Show primary keys, foreign keys, and application constraints - Not Started
   e. Indicate bitemporal validity columns where applicable - Not Started
   f. Generate diagram image or ensure it can be rendered in markdown - Not Started

7. Update architecture documentation - Not Started
   a. Add ER diagram section to architecture documentation or create separate database_schema.md file - Not Started
   b. Update architecture toc.md to include the new documentation - Not Started
   c. Add cross-references between architecture docs and SQL files - Not Started
   d. Verify all links work correctly - Not Started

8. Validation - Not Started
   a. Re-run database initialization to ensure commented SQL files still parse correctly - Not Started
   b. Verify all existing tests still pass - Not Started
   c. Review documentation for accuracy and completeness - Not Started

## Notes

### Documentation Format

SQL comments should use the following format:

```sql
-- TABLE-LEVEL COMMENT
-- Brief description of the table's purpose and role in the system.

CREATE TABLE example_table (
    -- column_name: Brief description of what this column represents
    --   and any important constraints or relationships.
    column_name    DATA_TYPE    CONSTRAINTS
);
```

For complex relationships that span multiple lines:

```sql
/*
  RELATIONSHIP: table_name -> other_table
  Description of the relationship and any application-level constraints
  that cannot be expressed as database foreign keys.
*/
```

### ER Diagram Tool Considerations

The project already uses D2 for some diagrams (see `../../src/backend/schema/overview.d2` and other `.d2` files). Consider using D2 for consistency, but evaluate other options:

| Tool         | Pros                                        | Cons                                 |
| ------------ | ------------------------------------------- | ------------------------------------ |
| D2           | Already in use, text-based, version control | Limited ER-specific features         |
| Mermaid      | Native GitHub support, widely known         | Limited styling options              |
| dbdiagram.io | Purpose-built for ER, intuitive             | External service, export needed      |
| DrawSQL      | Visual editor, good export options          | External service, manual sync needed |

### Tables to Document

| Schema Section | Tables                                                                                                                           |
| -------------- | -------------------------------------------------------------------------------------------------------------------------------- |
| Policy         | `namespace_family_type`, `namespace_type`, `subject_gateway`, `namespace_family_namespace_type`, `namespace_family`, `namespace` |
| Taxonomy       | `category`, `category_attr`, `element`, `element_attr`                                                                           |
| Timeseries     | `ts_doc_type`, `ts_doc_row_type`, `ts_doc`, `ts_doc_tags`, `ts_doc_input_lineage`                                                |
| Connectors     | `storage_config`                                                                                                                 |
| Morphisms      | `metric`, `metric_dim`, `transform`, `transform_input`, `transform_output`                                                       |

### Bitemporal Validity

Many tables include a `validity` column of type `tstzrange` that implements bitemporal versioning. Documentation should explain:

- The validity range represents when a version of an entity is valid
- Lower bound is inclusive, upper bound is exclusive (`[lower, upper)`)
- Updates use a "close and insert" pattern rather than in-place updates
- Some tables have validity disabled (see comments in those files)

### Open Questions

1. Should the ER diagram show both enabled and disabled validity columns?
2. Should application constraints (those that cannot be expressed as FKs) be shown differently in the diagram?
3. What level of detail is appropriate for the architecture documentation vs. the SQL files themselves?
