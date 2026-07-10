/*
  CHRONOCONE CONSOLIDATED SCHEMA

  ARCHITECTURAL REFERENCE:
  This schema is bitemporal. Mutations flow exclusively through
  ExecuteBitemporalMutation (EBM) in serde/rdbms/mutation.go,
  which implements the Flush/Purge/Assert pipeline and owns all
  interval algebra. Application code outside that file may emit
  only INSERT statements against bitemporal tables.

  Full documentation:
    memory-bank/specifications/architecture/database_design.md

  QUICK RULES:
    1. No UPDATEs or DELETEs against bitemporal tables outside
       serde/rdbms/mutation.go.
    2. Identity: override identity columns to reuse stable IDs
       across versions (OVERRIDING SYSTEM VALUE).
    3. Localization: human-facing strings reside in *_localization
       tables.
    4. Keys: technical keys (CITEXT) are 6-64 chars and stable
       across all versions of an entity.
    5. Audit: every row must include insertion_metadata
       (AuditContext).

  FORBIDDEN IDENTIFIERS (do not name a table, column, or alias
  these):
    values   -- collides with overridingInsert's SQL pivot
                (see serde/rdbms/mutation.go)
*/

CREATE EXTENSION IF NOT EXISTS citext;
CREATE EXTENSION IF NOT EXISTS btree_gist;

CREATE TYPE bitemporal_mode AS ENUM ('active', 'historical');

-- Create a dedicated schema for physical partitions to hide them from the application model (Go-Jet)
CREATE SCHEMA IF NOT EXISTS internal_partitions;

CREATE OR REPLACE FUNCTION create_bitemporal_partitions(table_name TEXT, entity_id_columns TEXT[])
RETURNS VOID AS $$
DECLARE
    entity_cols_str TEXT;
    entity_cols_with_equals TEXT;
    num_partitions INTEGER;
    i INTEGER;
BEGIN
    entity_cols_str := array_to_string(entity_id_columns, ', ');
    
    -- Determine the number of historical partitions.
    -- To override this (e.g. for production), run: SET chronocone.historical_partitions = '32';
    num_partitions := coalesce(current_setting('chronocone.historical_partitions', true)::integer, 2);

    -- Create active partition
    EXECUTE format('CREATE TABLE internal_partitions.%I_active PARTITION OF %I FOR VALUES IN (''active'')', table_name, table_name);
    -- Standard B-tree index
    EXECUTE format('CREATE INDEX %I_active_idx ON internal_partitions.%I_active (%s, validity_start DESC)', table_name, table_name, entity_cols_str);
    
    -- Create historical parent partition
    EXECUTE format('CREATE TABLE internal_partitions.%I_historical PARTITION OF %I FOR VALUES IN (''historical'') PARTITION BY HASH (%s)', table_name, table_name, entity_cols_str);
    
    -- Build entity_cols_with_equals for EXCLUDE constraint
    SELECT string_agg('(' || quote_ident(col) || '::text) WITH =', ', ') INTO entity_cols_with_equals
    FROM unnest(entity_id_columns) col;

    -- Create physical sub-partitions for historical data
    FOR i IN 0..(num_partitions - 1) LOOP
        EXECUTE format('CREATE TABLE internal_partitions.%I_historical_%s PARTITION OF internal_partitions.%I_historical FOR VALUES WITH (MODULUS %s, REMAINDER %s)', 
            table_name, i, table_name, num_partitions, i);
        -- Bitemporal exclusion constraint
        EXECUTE format('ALTER TABLE internal_partitions.%I_historical_%s ADD CONSTRAINT %I_historical_%s_exclude EXCLUDE USING gist (%s, tstzrange(validity_start, validity_end, ''[)'') WITH &&, tstzrange(transaction_start, transaction_end, ''[)'') WITH &&)', 
            table_name, i, table_name, i, entity_cols_with_equals);
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- SECTION 1: CONNECTORS & CORE
-- ============================================================================

CREATE TABLE language (
    language                    CITEXT       NOT NULL CHECK (length(language) BETWEEN 2 AND 3)
  , country                     CITEXT       NOT NULL CHECK (length(country) = 0 OR length(country) BETWEEN 2 AND 3)
  , validity_start              timestamptz  NOT NULL DEFAULT transaction_timestamp()
  , validity_end                timestamptz  NOT NULL DEFAULT '3000-01-01 00:00:00+00'::timestamptz CHECK (validity_start < validity_end)
  , transaction_start           timestamptz  NOT NULL DEFAULT transaction_timestamp()
  , transaction_end             timestamptz  NOT NULL DEFAULT '3000-01-01 00:00:00+00'::timestamptz
  , mode                        bitemporal_mode NOT NULL DEFAULT 'active'
  , display_name                TEXT         NOT NULL CHECK (length(display_name) BETWEEN 3 AND 64)
  , insertion_metadata          JSONB        NOT NULL
  , CONSTRAINT language_physical_pk          PRIMARY KEY (language, country, validity_end, transaction_end, mode)
) PARTITION BY LIST (mode);

SELECT create_bitemporal_partitions('language', ARRAY['language', 'country']);

INSERT INTO language (language, country, display_name, insertion_metadata)
VALUES
    ('en', '',   'English',         '{"reason": "initial seed"}')
  , ('en', 'US', 'English (US)',    '{"reason": "initial seed"}')
  , ('en', 'GB', 'English (UK)',    '{"reason": "initial seed"}')
;

CREATE TABLE storage_config (
    storage_config_id           INTEGER      GENERATED ALWAYS AS IDENTITY
  , validity_start              timestamptz  NOT NULL DEFAULT transaction_timestamp()
  , validity_end                timestamptz  NOT NULL DEFAULT '3000-01-01 00:00:00+00'::timestamptz CHECK (validity_start < validity_end)
  , transaction_start           timestamptz  NOT NULL DEFAULT transaction_timestamp()
  , transaction_end             timestamptz  NOT NULL DEFAULT '3000-01-01 00:00:00+00'::timestamptz
  , mode                        bitemporal_mode NOT NULL DEFAULT 'active'
  , key                         CITEXT       NOT NULL CHECK (length(key) BETWEEN 6 AND 64)
  , driver_name                 TEXT         NOT NULL -- this much match an enumerated choice from code
  , location                    TEXT         NOT NULL -- function parameter
  , credential_path             TEXT         NOT NULL -- function parameter
  , direction                   TEXT         NOT NULL CHECK (direction IN ('R', 'W', 'RW'))
  , insertion_metadata          JSONB        NOT NULL
  , CONSTRAINT storage_config_physical_pk    PRIMARY KEY (storage_config_id, validity_end, transaction_end, mode)
) PARTITION BY LIST (mode);

SELECT create_bitemporal_partitions('storage_config', ARRAY['storage_config_id']);

CREATE TABLE storage_config_localization (
    storage_config_id           INTEGER      NOT NULL
  , validity_start              timestamptz  NOT NULL DEFAULT transaction_timestamp()
  , validity_end                timestamptz  NOT NULL DEFAULT '3000-01-01 00:00:00+00'::timestamptz CHECK (validity_start < validity_end)
  , transaction_start           timestamptz  NOT NULL DEFAULT transaction_timestamp()
  , transaction_end             timestamptz  NOT NULL DEFAULT '3000-01-01 00:00:00+00'::timestamptz
  , mode                        bitemporal_mode NOT NULL DEFAULT 'active'
  , language                    CITEXT       NOT NULL CHECK (length(language) BETWEEN 2 AND 3)
  , country                     CITEXT       NOT NULL CHECK (length(country) = 0 OR length(country) BETWEEN 2 AND 3)
  , display_name                TEXT         NOT NULL CHECK (length(display_name) BETWEEN 6 AND 64)
  , tooltip                     TEXT         NOT NULL CHECK (length(tooltip) BETWEEN 15 AND 256)
  , description                 TEXT         NOT NULL CHECK (length(description) > 25)
  , insertion_metadata          JSONB        NOT NULL
  , CONSTRAINT storage_config_localization_physical_pk     PRIMARY KEY (storage_config_id, language, country, validity_end, transaction_end, mode)
) PARTITION BY LIST (mode);

SELECT create_bitemporal_partitions('storage_config_localization', ARRAY['storage_config_id', 'language', 'country']);

-- ============================================================================
-- SECTION 2: POLICY & IDENTITY
-- ============================================================================

CREATE TABLE namespace_family_type (
    ns_fam_type_id              SMALLINT     NOT NULL  PRIMARY KEY
  , key                         citext       NOT NULL  UNIQUE       CHECK (LENGTH(key) BETWEEN 6 AND 64)
)
;

INSERT INTO namespace_family_type
    (ns_fam_type_id, key)
VALUES
    (1, 'application')
  , (2, 'personal')
  , (3, 'department')
;

CREATE TABLE namespace_family_type_localization (
    ns_fam_type_id              SMALLINT     NOT NULL
  , validity_start              timestamptz  NOT NULL DEFAULT transaction_timestamp()
  , validity_end                timestamptz  NOT NULL DEFAULT '3000-01-01 00:00:00+00'::timestamptz CHECK (validity_start < validity_end)
  , transaction_start           timestamptz  NOT NULL DEFAULT transaction_timestamp()
  , transaction_end             timestamptz  NOT NULL DEFAULT '3000-01-01 00:00:00+00'::timestamptz
  , mode                        bitemporal_mode NOT NULL DEFAULT 'active'
  , language                    CITEXT       NOT NULL CHECK (length(language) BETWEEN 2 AND 3)
  , country                     CITEXT       NOT NULL CHECK (length(country) = 0 OR length(country) BETWEEN 2 AND 3)
  , display_name                TEXT         NOT NULL CHECK (length(display_name) BETWEEN 6 AND 64)
  , tooltip                     TEXT         NOT NULL CHECK (length(tooltip) BETWEEN 15 AND 256)
  , description                 TEXT         NOT NULL CHECK (length(description) > 25)
  , insertion_metadata          JSONB        NOT NULL
  , CONSTRAINT ns_fam_type_localization_physical_pk PRIMARY KEY (ns_fam_type_id, language, country, validity_end, transaction_end, mode)
) PARTITION BY LIST (mode);

SELECT create_bitemporal_partitions('namespace_family_type_localization', ARRAY['ns_fam_type_id', 'language', 'country']);

CREATE TABLE namespace_type (
    ns_type_id                  SMALLINT     NOT NULL  PRIMARY KEY
  , key                         citext       NOT NULL  UNIQUE       CHECK (LENGTH(key) BETWEEN 6 AND 64)
)
;

INSERT INTO namespace_type
    (ns_type_id, key)
VALUES
    (1, 'Production')
  , (2, 'Smoke Test')
  , (3, 'Integration')
  , (4, 'Acceptance (UAT)')
  , (5, 'Alpha Test')
  , (6, 'Scenario')
  , (7, 'Experiment')
  , (8, 'Beta Test')
;

CREATE TABLE namespace_type_localization (
    ns_type_id                  SMALLINT     NOT NULL
  , validity_start              timestamptz  NOT NULL DEFAULT transaction_timestamp()
  , validity_end                timestamptz  NOT NULL DEFAULT '3000-01-01 00:00:00+00'::timestamptz CHECK (validity_start < validity_end)
  , transaction_start           timestamptz  NOT NULL DEFAULT transaction_timestamp()
  , transaction_end             timestamptz  NOT NULL DEFAULT '3000-01-01 00:00:00+00'::timestamptz
  , mode                        bitemporal_mode NOT NULL DEFAULT 'active'
  , language                    CITEXT       NOT NULL CHECK (length(language) BETWEEN 2 AND 3)
  , country                     CITEXT       NOT NULL CHECK (length(country) = 0 OR length(country) BETWEEN 2 AND 3)
  , display_name                TEXT         NOT NULL CHECK (length(display_name) BETWEEN 6 AND 64)
  , tooltip                     TEXT         NOT NULL CHECK (length(tooltip) BETWEEN 15 AND 256)
  , description                 TEXT         NOT NULL CHECK (length(description) > 25)
  , insertion_metadata          JSONB        NOT NULL
  , CONSTRAINT ns_type_localization_physical_pk PRIMARY KEY (ns_type_id, language, country, validity_end, transaction_end, mode)
) PARTITION BY LIST (mode);

SELECT create_bitemporal_partitions('namespace_type_localization', ARRAY['ns_type_id', 'language', 'country']);

CREATE TABLE subject_gateway (
    subject_id                  INTEGER      GENERATED ALWAYS AS IDENTITY PRIMARY KEY
  , oidc_provider               TEXT         NOT NULL COLLATE "C" -- FK: OIDC provider
  , oidc_subject                TEXT         NOT NULL COLLATE "C" -- FK: subject inside OIDC provider
  , CONSTRAINT subject_logical_pk
        UNIQUE (oidc_provider, oidc_subject)
)
;

INSERT INTO subject_gateway
    (subject_id, oidc_provider, oidc_subject)
OVERRIDING SYSTEM VALUE VALUES
    (-1, '', 'nobody')
;

CREATE TABLE namespace_family_namespace_type (
    customer_subject_id          INTEGER      NOT NULL  REFERENCES subject_gateway
  , ns_fam_type_id               SMALLINT     NOT NULL  REFERENCES namespace_family_type
  , ns_type_id                   SMALLINT     NOT NULL  REFERENCES namespace_type
  , PRIMARY KEY (customer_subject_id, ns_fam_type_id, ns_type_id)
  , owner_can_delete_family      BOOLEAN      NOT NULL
  , ns_can_map_outside_fam       BOOLEAN      NOT NULL
  , ns_can_be_mapped_outside_fam BOOLEAN      NOT NULL
  , owner_can_delete_ns          BOOLEAN      NOT NULL
  , ns_can_map_others            BOOLEAN      NOT NULL
  , ns_can_be_mapped             BOOLEAN      NOT NULL
  , shortcut_name                VARCHAR      NOT NULL
)
;

INSERT INTO namespace_family_namespace_type 
    (customer_subject_id, ns_fam_type_id, ns_type_id, owner_can_delete_family, ns_can_map_outside_fam, ns_can_be_mapped_outside_fam, owner_can_delete_ns, ns_can_map_others, ns_can_be_mapped, shortcut_name)
VALUES
    -- Application lifecycle
    (-1, 1, 1, FALSE, FALSE, TRUE,  FALSE, FALSE, TRUE,  'Production application: Plan-of-record')
  , (-1, 1, 8, FALSE, TRUE,  FALSE, FALSE, TRUE,  FALSE, 'Production application: Public Beta')
  , (-1, 1, 2, TRUE,  FALSE, FALSE, TRUE,  FALSE, FALSE, 'Production application: Deployment Smoke Test')
  , (-1, 1, 3, TRUE,  FALSE, TRUE,  FALSE, FALSE, TRUE,  'Production application: Deployment Integration Test')
  , (-1, 1, 4, TRUE,  FALSE, FALSE, TRUE,  FALSE, FALSE, 'Production application: User Acceptance Test')
  , (-1, 1, 5, TRUE,  FALSE, TRUE,  FALSE, FALSE, TRUE,  'Production application: Developer Alpha Environment')
  , (-1, 1, 6, TRUE,  FALSE, FALSE, TRUE,  FALSE, FALSE, 'Production application: Scenario')
    -- Personal uses
  , (-1, 2, 7, TRUE,  TRUE,  FALSE, TRUE,  TRUE,  FALSE, 'Personal Sandbox (Experiment)')
  , (-1, 2, 6, TRUE,  TRUE,  FALSE, TRUE,  TRUE,  TRUE,  'Personal Lab (Scenario)')
    -- Team namespaces prioritize shared resources and collaborative mapping.
  , (-1, 3, 1, FALSE, TRUE,  TRUE,  FALSE, TRUE,  TRUE,  'Team: Plan-of-record')
  , (-1, 3, 8, FALSE, TRUE,  TRUE,  FALSE, TRUE,  TRUE,  'Team: Beta')
  , (-1, 3, 5, TRUE,  TRUE,  TRUE,  FALSE, TRUE,  TRUE,  'Team: Alpha')
  , (-1, 3, 6, TRUE,  TRUE,  TRUE,  TRUE,  TRUE,  TRUE,  'Team: Scenario')
  , (-1, 3, 7, TRUE,  TRUE,  TRUE,  TRUE,  TRUE,  TRUE,  'Team: Experiment')
;

CREATE TABLE namespace_family (
    ns_fam_id                   INTEGER      NOT NULL  GENERATED ALWAYS AS IDENTITY
  , validity_start              timestamptz  NOT NULL  DEFAULT transaction_timestamp()
  , validity_end                timestamptz  NOT NULL  DEFAULT '3000-01-01 00:00:00+00'::timestamptz CHECK (validity_start < validity_end)
  , transaction_start           timestamptz  NOT NULL DEFAULT transaction_timestamp()
  , transaction_end             timestamptz  NOT NULL DEFAULT '3000-01-01 00:00:00+00'::timestamptz
  , mode                        bitemporal_mode NOT NULL DEFAULT 'active'
  , customer_subject_id         INTEGER      NOT NULL
  , owner_subject_id            INTEGER      NOT NULL
  , ns_fam_type_id              SMALLINT     NOT NULL  REFERENCES namespace_family_type
  , key                         citext       NOT NULL  CHECK (LENGTH(key) BETWEEN 6 AND 64)
  , insertion_metadata          JSONB        NOT NULL
  , CONSTRAINT ns_fam_physical_pk   PRIMARY KEY (ns_fam_id, validity_end, transaction_end, mode)
  , CONSTRAINT ns_fam_customer_fk   FOREIGN KEY (customer_subject_id) REFERENCES subject_gateway(subject_id)
  , CONSTRAINT ns_fam_owner_fk      FOREIGN KEY (owner_subject_id) REFERENCES subject_gateway(subject_id)
) PARTITION BY LIST (mode);

SELECT create_bitemporal_partitions('namespace_family', ARRAY['ns_fam_id']);

CREATE TABLE namespace_family_localization (
    ns_fam_id                   INTEGER      NOT NULL
  , validity_start              timestamptz  NOT NULL DEFAULT transaction_timestamp()
  , validity_end                timestamptz  NOT NULL DEFAULT '3000-01-01 00:00:00+00'::timestamptz CHECK (validity_start < validity_end)
  , transaction_start           timestamptz  NOT NULL DEFAULT transaction_timestamp()
  , transaction_end             timestamptz  NOT NULL DEFAULT '3000-01-01 00:00:00+00'::timestamptz
  , mode                        bitemporal_mode NOT NULL DEFAULT 'active'
  , language                    CITEXT       NOT NULL CHECK (length(language) BETWEEN 2 AND 3)
  , country                     CITEXT       NOT NULL CHECK (length(country) = 0 OR length(country) BETWEEN 2 AND 3)
  , display_name                TEXT         NOT NULL CHECK (length(display_name) BETWEEN 6 AND 64)
  , tooltip                     TEXT         NOT NULL CHECK (length(tooltip) BETWEEN 15 AND 256)
  , description                 TEXT         NOT NULL CHECK (length(description) > 25)
  , insertion_metadata          JSONB        NOT NULL
  , CONSTRAINT ns_fam_localization_physical_pk PRIMARY KEY (ns_fam_id, language, country, validity_end, transaction_end, mode)
) PARTITION BY LIST (mode);

SELECT create_bitemporal_partitions('namespace_family_localization', ARRAY['ns_fam_id', 'language', 'country']);

CREATE TABLE namespace (
    ns_id                       BIGINT       NOT NULL  GENERATED ALWAYS AS IDENTITY
  , validity_start              timestamptz  NOT NULL  DEFAULT transaction_timestamp()
  , validity_end                timestamptz  NOT NULL  DEFAULT '3000-01-01 00:00:00+00'::timestamptz CHECK (validity_start < validity_end)
  , transaction_start           timestamptz  NOT NULL DEFAULT transaction_timestamp()
  , transaction_end             timestamptz  NOT NULL DEFAULT '3000-01-01 00:00:00+00'::timestamptz
  , mode                        bitemporal_mode NOT NULL DEFAULT 'active'
  , owner_subject_id            INTEGER      NOT NULL
  , ns_fam_id                   INTEGER      NOT NULL
  , ns_type_id                  SMALLINT     NOT NULL  REFERENCES namespace_type
  , key                         citext       NOT NULL  CHECK (LENGTH(key) BETWEEN 6 AND 64)
  , insertion_metadata          JSONB        NOT NULL
  , CONSTRAINT ns_physical_pk   PRIMARY KEY (ns_id, validity_end, transaction_end, mode)
  , CONSTRAINT ns_owner_fk      FOREIGN KEY (owner_subject_id) REFERENCES subject_gateway(subject_id)
) PARTITION BY LIST (mode);

SELECT create_bitemporal_partitions('namespace', ARRAY['ns_id']);

CREATE TABLE namespace_localization (
    ns_id                       BIGINT       NOT NULL
  , validity_start              timestamptz  NOT NULL DEFAULT transaction_timestamp()
  , validity_end                timestamptz  NOT NULL DEFAULT '3000-01-01 00:00:00+00'::timestamptz CHECK (validity_start < validity_end)
  , transaction_start           timestamptz  NOT NULL DEFAULT transaction_timestamp()
  , transaction_end             timestamptz  NOT NULL DEFAULT '3000-01-01 00:00:00+00'::timestamptz
  , mode                        bitemporal_mode NOT NULL DEFAULT 'active'
  , language                    CITEXT       NOT NULL CHECK (length(language) BETWEEN 2 AND 3)
  , country                     CITEXT       NOT NULL CHECK (length(country) = 0 OR length(country) BETWEEN 2 AND 3)
  , display_name                TEXT         NOT NULL CHECK (length(display_name) BETWEEN 6 AND 64)
  , tooltip                     TEXT         NOT NULL CHECK (length(tooltip) BETWEEN 15 AND 256)
  , description                 TEXT         NOT NULL CHECK (length(description) > 25)
  , insertion_metadata          JSONB        NOT NULL
  , CONSTRAINT ns_localization_physical_pk PRIMARY KEY (ns_id, language, country, validity_end, transaction_end, mode)
) PARTITION BY LIST (mode);

SELECT create_bitemporal_partitions('namespace_localization', ARRAY['ns_id', 'language', 'country']);

-- ============================================================================
-- SECTION 3: TAXONOMY
-- ============================================================================

CREATE TABLE category (
    category_id                 BIGINT       GENERATED ALWAYS AS IDENTITY
  , ns_id                       BIGINT       NOT NULL
  , validity_start              timestamptz  NOT NULL DEFAULT transaction_timestamp()
  , validity_end                timestamptz  NOT NULL DEFAULT '3000-01-01 00:00:00+00'::timestamptz CHECK (validity_start < validity_end)
  , transaction_start           timestamptz  NOT NULL DEFAULT transaction_timestamp()
  , transaction_end             timestamptz  NOT NULL DEFAULT '3000-01-01 00:00:00+00'::timestamptz
  , mode                        bitemporal_mode NOT NULL DEFAULT 'active'
  , key                         citext       NOT NULL CHECK (LENGTH(key) BETWEEN 6 AND 64)
  , insertion_metadata          JSONB        NOT NULL
  , CONSTRAINT category_physical_pk            PRIMARY KEY (category_id, validity_end, transaction_end, mode)
) PARTITION BY LIST (mode);

SELECT create_bitemporal_partitions('category', ARRAY['category_id']);

CREATE TABLE category_localization (
    category_id                 BIGINT       NOT NULL
  , validity_start              timestamptz  NOT NULL DEFAULT transaction_timestamp()
  , validity_end                timestamptz  NOT NULL DEFAULT '3000-01-01 00:00:00+00'::timestamptz CHECK (validity_start < validity_end)
  , transaction_start           timestamptz  NOT NULL DEFAULT transaction_timestamp()
  , transaction_end             timestamptz  NOT NULL DEFAULT '3000-01-01 00:00:00+00'::timestamptz
  , mode                        bitemporal_mode NOT NULL DEFAULT 'active'
  , language                    CITEXT       NOT NULL CHECK (length(language) BETWEEN 2 AND 3)
  , country                     CITEXT       NOT NULL CHECK (length(country) = 0 OR length(country) BETWEEN 2 AND 3)
  , display_name                TEXT         NOT NULL CHECK (length(display_name) BETWEEN 6 AND 64)
  , tooltip                     TEXT         NOT NULL CHECK (length(tooltip) BETWEEN 15 AND 256)
  , description                 TEXT         NOT NULL CHECK (length(description) > 25)
  , insertion_metadata          JSONB        NOT NULL
  , CONSTRAINT category_localization_physical_pk PRIMARY KEY (category_id, language, country, validity_end, transaction_end, mode)
) PARTITION BY LIST (mode);

SELECT create_bitemporal_partitions('category_localization', ARRAY['category_id', 'language', 'country']);

CREATE TABLE category_shape (
    category_shape_id           BIGINT       NOT NULL GENERATED ALWAYS AS IDENTITY
  , category_id                 BIGINT       NOT NULL
  , asserting_ns_id             BIGINT       NOT NULL
  , validity_start              timestamptz  NOT NULL DEFAULT transaction_timestamp()
  , validity_end                timestamptz  NOT NULL DEFAULT '3000-01-01 00:00:00+00'::timestamptz CHECK (validity_start < validity_end)
  , transaction_start           timestamptz  NOT NULL DEFAULT transaction_timestamp()
  , transaction_end             timestamptz  NOT NULL DEFAULT '3000-01-01 00:00:00+00'::timestamptz
  , mode                        bitemporal_mode NOT NULL DEFAULT 'active'
  , key                         citext       NOT NULL CHECK (LENGTH(key) BETWEEN 6 AND 64)
  , is_required                 BOOLEAN      NOT NULL DEFAULT FALSE
  , json_schema                 JSONB        NOT NULL
  , insertion_metadata          JSONB        NOT NULL
  , CONSTRAINT category_shape_physical_pk    PRIMARY KEY (category_shape_id, validity_end, transaction_end, mode)
) PARTITION BY LIST (mode);

SELECT create_bitemporal_partitions('category_shape', ARRAY['category_shape_id']);

CREATE TABLE category_shape_localization (
    category_shape_id           BIGINT       NOT NULL
  , validity_start              timestamptz  NOT NULL DEFAULT transaction_timestamp()
  , validity_end                timestamptz  NOT NULL DEFAULT '3000-01-01 00:00:00+00'::timestamptz CHECK (validity_start < validity_end)
  , transaction_start           timestamptz  NOT NULL DEFAULT transaction_timestamp()
  , transaction_end             timestamptz  NOT NULL DEFAULT '3000-01-01 00:00:00+00'::timestamptz
  , mode                        bitemporal_mode NOT NULL DEFAULT 'active'
  , language                    CITEXT       NOT NULL CHECK (length(language) BETWEEN 2 AND 3)
  , country                     CITEXT       NOT NULL CHECK (length(country) = 0 OR length(country) BETWEEN 2 AND 3)
  , display_name                TEXT         NOT NULL CHECK (length(display_name) BETWEEN 6 AND 64)
  , tooltip                     TEXT         NOT NULL CHECK (length(tooltip) BETWEEN 15 AND 256)
  , description                 TEXT         NOT NULL CHECK (length(description) > 25)
  , insertion_metadata          JSONB        NOT NULL
  , CONSTRAINT category_shape_localization_physical_pk PRIMARY KEY (category_shape_id, language, country, validity_end, transaction_end, mode)
) PARTITION BY LIST (mode);

SELECT create_bitemporal_partitions('category_shape_localization', ARRAY['category_shape_id', 'language', 'country']);

CREATE TABLE category_attr (
    category_attr_id            BIGINT       NOT NULL GENERATED ALWAYS AS IDENTITY
  , category_id                 BIGINT       NOT NULL
  , attr_owner_ns_id            BIGINT       NOT NULL
  , validity_start              timestamptz  NOT NULL DEFAULT transaction_timestamp()
  , validity_end                timestamptz  NOT NULL DEFAULT '3000-01-01 00:00:00+00'::timestamptz CHECK (validity_start < validity_end)
  , transaction_start           timestamptz  NOT NULL DEFAULT transaction_timestamp()
  , transaction_end             timestamptz  NOT NULL DEFAULT '3000-01-01 00:00:00+00'::timestamptz
  , mode                        bitemporal_mode NOT NULL DEFAULT 'active'
  , key                         citext       NOT NULL CHECK (LENGTH(key) BETWEEN 6 AND 64)
  , value                       JSONB        NOT NULL
  , insertion_metadata          JSONB        NOT NULL
  , CONSTRAINT category_attr_physical_pk  PRIMARY KEY (category_attr_id, validity_end, transaction_end, mode)
) PARTITION BY LIST (mode);

SELECT create_bitemporal_partitions('category_attr', ARRAY['category_attr_id']);

CREATE TABLE category_attr_localization (
    category_attr_id            BIGINT       NOT NULL
  , validity_start              timestamptz  NOT NULL DEFAULT transaction_timestamp()
  , validity_end                timestamptz  NOT NULL DEFAULT '3000-01-01 00:00:00+00'::timestamptz CHECK (validity_start < validity_end)
  , transaction_start           timestamptz  NOT NULL DEFAULT transaction_timestamp()
  , transaction_end             timestamptz  NOT NULL DEFAULT '3000-01-01 00:00:00+00'::timestamptz
  , mode                        bitemporal_mode NOT NULL DEFAULT 'active'
  , language                    CITEXT       NOT NULL CHECK (length(language) BETWEEN 2 AND 3)
  , country                     CITEXT       NOT NULL CHECK (length(country) = 0 OR length(country) BETWEEN 2 AND 3)
  , display_name                TEXT         NOT NULL CHECK (length(display_name) BETWEEN 6 AND 64)
  , tooltip                     TEXT         NOT NULL CHECK (length(tooltip) BETWEEN 15 AND 256)
  , description                 TEXT         NOT NULL CHECK (length(description) > 25)
  , insertion_metadata          JSONB        NOT NULL
  , CONSTRAINT category_attr_localization_physical_pk PRIMARY KEY (category_attr_id, language, country, validity_end, transaction_end, mode)
) PARTITION BY LIST (mode);

SELECT create_bitemporal_partitions('category_attr_localization', ARRAY['category_attr_id', 'language', 'country']);

CREATE TABLE element (
    element_id                  BIGINT       GENERATED ALWAYS AS IDENTITY
  , category_id                 BIGINT       NOT NULL
  , validity_start              timestamptz  NOT NULL DEFAULT transaction_timestamp()
  , validity_end                timestamptz  NOT NULL DEFAULT '3000-01-01 00:00:00+00'::timestamptz CHECK (validity_start < validity_end)
  , transaction_start           timestamptz  NOT NULL DEFAULT transaction_timestamp()
  , transaction_end             timestamptz  NOT NULL DEFAULT '3000-01-01 00:00:00+00'::timestamptz
  , mode                        bitemporal_mode NOT NULL DEFAULT 'active'
  , key                         citext       NOT NULL CHECK (LENGTH(key) BETWEEN 6 AND 64)
  , insertion_metadata          JSONB        NOT NULL
  , CONSTRAINT element_physical_pk   PRIMARY KEY (element_id, validity_end, transaction_end, mode)
) PARTITION BY LIST (mode);

SELECT create_bitemporal_partitions('element', ARRAY['element_id']);

CREATE TABLE element_localization (
    element_id                  BIGINT       NOT NULL
  , validity_start              timestamptz  NOT NULL DEFAULT transaction_timestamp()
  , validity_end                timestamptz  NOT NULL DEFAULT '3000-01-01 00:00:00+00'::timestamptz CHECK (validity_start < validity_end)
  , transaction_start           timestamptz  NOT NULL DEFAULT transaction_timestamp()
  , transaction_end             timestamptz  NOT NULL DEFAULT '3000-01-01 00:00:00+00'::timestamptz
  , mode                        bitemporal_mode NOT NULL DEFAULT 'active'
  , language                    CITEXT       NOT NULL CHECK (length(language) BETWEEN 2 AND 3)
  , country                     CITEXT       NOT NULL CHECK (length(country) = 0 OR length(country) BETWEEN 2 AND 3)
  , display_name                TEXT         NOT NULL CHECK (length(display_name) BETWEEN 6 AND 64)
  , tooltip                     TEXT         NOT NULL CHECK (length(tooltip) BETWEEN 15 AND 256)
  , description                 TEXT         NOT NULL CHECK (length(description) > 25)
  , insertion_metadata          JSONB        NOT NULL
  , CONSTRAINT element_localization_physical_pk PRIMARY KEY (element_id, language, country, validity_end, transaction_end, mode)
) PARTITION BY LIST (mode);

SELECT create_bitemporal_partitions('element_localization', ARRAY['element_id', 'language', 'country']);

CREATE TABLE element_attr (
    element_attr_id             BIGINT       NOT NULL GENERATED ALWAYS AS IDENTITY
  , element_id                  BIGINT       NOT NULL
  , category_shape_id           BIGINT       NOT NULL
  , validity_start              timestamptz  NOT NULL DEFAULT transaction_timestamp()
  , validity_end                timestamptz  NOT NULL DEFAULT '3000-01-01 00:00:00+00'::timestamptz CHECK (validity_start < validity_end)
  , transaction_start           timestamptz  NOT NULL DEFAULT transaction_timestamp()
  , transaction_end             timestamptz  NOT NULL DEFAULT '3000-01-01 00:00:00+00'::timestamptz
  , mode                        bitemporal_mode NOT NULL DEFAULT 'active'
  , value                       JSONB        NOT NULL
  , insertion_metadata          JSONB        NOT NULL
  , CONSTRAINT element_attr_physical_pk      PRIMARY KEY (element_attr_id, validity_end, transaction_end, mode)
) PARTITION BY LIST (mode);

SELECT create_bitemporal_partitions('element_attr', ARRAY['element_attr_id']);

CREATE TABLE element_attr_localization (
    element_attr_id             BIGINT       NOT NULL
  , validity_start              timestamptz  NOT NULL DEFAULT transaction_timestamp()
  , validity_end                timestamptz  NOT NULL DEFAULT '3000-01-01 00:00:00+00'::timestamptz CHECK (validity_start < validity_end)
  , transaction_start           timestamptz  NOT NULL DEFAULT transaction_timestamp()
  , transaction_end             timestamptz  NOT NULL DEFAULT '3000-01-01 00:00:00+00'::timestamptz
  , mode                        bitemporal_mode NOT NULL DEFAULT 'active'
  , language                    CITEXT       NOT NULL CHECK (length(language) BETWEEN 2 AND 3)
  , country                     CITEXT       NOT NULL CHECK (length(country) = 0 OR length(country) BETWEEN 2 AND 3)
  , display_name                TEXT         NOT NULL CHECK (length(display_name) BETWEEN 6 AND 64)
  , tooltip                     TEXT         NOT NULL CHECK (length(tooltip) BETWEEN 15 AND 256)
  , description                 TEXT         NOT NULL CHECK (length(description) > 25)
  , insertion_metadata          JSONB        NOT NULL
  , CONSTRAINT element_attr_localization_physical_pk PRIMARY KEY (element_attr_id, language, country, validity_end, transaction_end, mode)
) PARTITION BY LIST (mode);

SELECT create_bitemporal_partitions('element_attr_localization', ARRAY['element_attr_id', 'language', 'country']);

-- ============================================================================
-- SECTION 4: MORPHISMS
-- ============================================================================

CREATE TABLE metric (
    metric_id                   BIGINT       NOT NULL  GENERATED ALWAYS AS IDENTITY
  , ns_id                       BIGINT       NOT NULL
  , concept_id                  BIGINT       NOT NULL
  , validity_start              timestamptz  NOT NULL DEFAULT transaction_timestamp()
  , validity_end                timestamptz  NOT NULL DEFAULT '3000-01-01 00:00:00+00'::timestamptz CHECK (validity_start < validity_end)
  , transaction_start           timestamptz  NOT NULL DEFAULT transaction_timestamp()
  , transaction_end             timestamptz  NOT NULL DEFAULT '3000-01-01 00:00:00+00'::timestamptz
  , mode                        bitemporal_mode NOT NULL DEFAULT 'active'
  , key                         CITEXT       NOT NULL CHECK (length(key) BETWEEN 6 AND 64)
  , insertion_metadata          JSONB        NOT NULL
  , CONSTRAINT metric_physical_pk            PRIMARY KEY (metric_id, validity_end, transaction_end, mode)
) PARTITION BY LIST (mode);

SELECT create_bitemporal_partitions('metric', ARRAY['metric_id']);

CREATE TABLE metric_localization (
    metric_id                   BIGINT       NOT NULL
  , validity_start              timestamptz  NOT NULL DEFAULT transaction_timestamp()
  , validity_end                timestamptz  NOT NULL DEFAULT '3000-01-01 00:00:00+00'::timestamptz CHECK (validity_start < validity_end)
  , transaction_start           timestamptz  NOT NULL DEFAULT transaction_timestamp()
  , transaction_end             timestamptz  NOT NULL DEFAULT '3000-01-01 00:00:00+00'::timestamptz
  , mode                        bitemporal_mode NOT NULL DEFAULT 'active'
  , language                    CITEXT       NOT NULL CHECK (length(language) BETWEEN 2 AND 3)
  , country                     CITEXT       NOT NULL CHECK (length(country) = 0 OR length(country) BETWEEN 2 AND 3)
  , display_name                TEXT         NOT NULL CHECK (length(display_name) BETWEEN 6 AND 64)
  , tooltip                     TEXT         NOT NULL CHECK (length(tooltip) BETWEEN 15 AND 256)
  , description                 TEXT         NOT NULL CHECK (length(description) > 25)
  , insertion_metadata          JSONB        NOT NULL
  , CONSTRAINT metric_localization_physical_pk PRIMARY KEY (metric_id, language, country, validity_end, transaction_end, mode)
) PARTITION BY LIST (mode);

SELECT create_bitemporal_partitions('metric_localization', ARRAY['metric_id', 'language', 'country']);

CREATE TABLE metric_dim (
    metric_id                   BIGINT       NOT NULL
  , dimension_id                BIGINT       NOT NULL
  , validity_start              timestamptz  NOT NULL DEFAULT transaction_timestamp()
  , validity_end                timestamptz  NOT NULL DEFAULT '3000-01-01 00:00:00+00'::timestamptz CHECK (validity_start < validity_end)
  , transaction_start           timestamptz  NOT NULL DEFAULT transaction_timestamp()
  , transaction_end             timestamptz  NOT NULL DEFAULT '3000-01-01 00:00:00+00'::timestamptz
  , mode                        bitemporal_mode NOT NULL DEFAULT 'active'
  , insertion_metadata          JSONB        NOT NULL
  , dimension_order             SMALLINT     NOT NULL
  , CONSTRAINT metric_dim_physical_pk        PRIMARY KEY (metric_id, dimension_id, validity_end, transaction_end, mode)
) PARTITION BY LIST (mode);

SELECT create_bitemporal_partitions('metric_dim', ARRAY['metric_id', 'dimension_id']);

CREATE TABLE transformer (
    transformer_id              BIGINT       NOT NULL  GENERATED ALWAYS AS IDENTITY
  , validity_start              timestamptz  NOT NULL DEFAULT transaction_timestamp()
  , validity_end                timestamptz  NOT NULL DEFAULT '3000-01-01 00:00:00+00'::timestamptz CHECK (validity_start < validity_end)
  , transaction_start           timestamptz  NOT NULL DEFAULT transaction_timestamp()
  , transaction_end             timestamptz  NOT NULL DEFAULT '3000-01-01 00:00:00+00'::timestamptz
  , mode                        bitemporal_mode NOT NULL DEFAULT 'active'
  , ns_id                       BIGINT       NOT NULL
  , storage_config_id           INTEGER      NOT NULL
  , key                         CITEXT       NOT NULL CHECK (length(key) BETWEEN 6 AND 64)
  , code_location               TEXT         NOT NULL
  , insertion_metadata          JSONB        NOT NULL
  , CONSTRAINT transformer_physical_pk       PRIMARY KEY (transformer_id, validity_end, transaction_end, mode)
) PARTITION BY LIST (mode);

SELECT create_bitemporal_partitions('transformer', ARRAY['transformer_id']);

CREATE TABLE transformer_localization (
    transformer_id              BIGINT       NOT NULL
  , validity_start              timestamptz  NOT NULL DEFAULT transaction_timestamp()
  , validity_end                timestamptz  NOT NULL DEFAULT '3000-01-01 00:00:00+00'::timestamptz CHECK (validity_start < validity_end)
  , transaction_start           timestamptz  NOT NULL DEFAULT transaction_timestamp()
  , transaction_end             timestamptz  NOT NULL DEFAULT '3000-01-01 00:00:00+00'::timestamptz
  , mode                        bitemporal_mode NOT NULL DEFAULT 'active'
  , language                    CITEXT       NOT NULL CHECK (length(language) BETWEEN 2 AND 3)
  , country                     CITEXT       NOT NULL CHECK (length(country) = 0 OR length(country) BETWEEN 2 AND 3)
  , display_name                TEXT         NOT NULL CHECK (length(display_name) BETWEEN 6 AND 64)
  , tooltip                     TEXT         NOT NULL CHECK (length(tooltip) BETWEEN 15 AND 256)
  , description                 TEXT         NOT NULL CHECK (length(description) > 25)
  , insertion_metadata          JSONB        NOT NULL
  , CONSTRAINT transformer_localization_physical_pk PRIMARY KEY (transformer_id, language, country, validity_end, transaction_end, mode)
) PARTITION BY LIST (mode);

SELECT create_bitemporal_partitions('transformer_localization', ARRAY['transformer_id', 'language', 'country']);

CREATE TABLE transformer_input (
    transformer_id              BIGINT       NOT NULL
  , metric_id                   BIGINT       NOT NULL
  , validity_start              timestamptz  NOT NULL DEFAULT transaction_timestamp()
  , validity_end                timestamptz  NOT NULL DEFAULT '3000-01-01 00:00:00+00'::timestamptz CHECK (validity_start < validity_end)
  , transaction_start           timestamptz  NOT NULL DEFAULT transaction_timestamp()
  , transaction_end             timestamptz  NOT NULL DEFAULT '3000-01-01 00:00:00+00'::timestamptz
  , mode                        bitemporal_mode NOT NULL DEFAULT 'active'
  , input_order                 SMALLINT     NOT NULL
  , CONSTRAINT transformer_input_physical_pk   PRIMARY KEY (transformer_id, metric_id, validity_end, transaction_end, mode)
) PARTITION BY LIST (mode);

SELECT create_bitemporal_partitions('transformer_input', ARRAY['transformer_id', 'metric_id']);

CREATE TABLE transformer_output (
    transformer_id              BIGINT       NOT NULL
  , metric_id                   BIGINT       NOT NULL
  , validity_start              timestamptz  NOT NULL DEFAULT transaction_timestamp()
  , validity_end                timestamptz  NOT NULL DEFAULT '3000-01-01 00:00:00+00'::timestamptz CHECK (validity_start < validity_end)
  , transaction_start           timestamptz  NOT NULL DEFAULT transaction_timestamp()
  , transaction_end             timestamptz  NOT NULL DEFAULT '3000-01-01 00:00:00+00'::timestamptz
  , mode                        bitemporal_mode NOT NULL DEFAULT 'active'
  , CONSTRAINT transformer_output_physical_pk   PRIMARY KEY (transformer_id, metric_id, validity_end, transaction_end, mode)
) PARTITION BY LIST (mode);

SELECT create_bitemporal_partitions('transformer_output', ARRAY['transformer_id', 'metric_id']);

CREATE TABLE shifter (
    shifter_id                  BIGINT       NOT NULL  GENERATED ALWAYS AS IDENTITY
  , validity_start              timestamptz  NOT NULL DEFAULT transaction_timestamp()
  , validity_end                timestamptz  NOT NULL DEFAULT '3000-01-01 00:00:00+00'::timestamptz CHECK (validity_start < validity_end)
  , transaction_start           timestamptz  NOT NULL DEFAULT transaction_timestamp()
  , transaction_end             timestamptz  NOT NULL DEFAULT '3000-01-01 00:00:00+00'::timestamptz
  , mode                        bitemporal_mode NOT NULL DEFAULT 'active'
  , ns_id                       BIGINT       NOT NULL
  , storage_config_id           INTEGER      NOT NULL
  , key                         CITEXT       NOT NULL CHECK (length(key) BETWEEN 6 AND 64)
  , code_location               TEXT         NOT NULL
  , insertion_metadata          JSONB        NOT NULL
  , CONSTRAINT shifter_physical_pk       PRIMARY KEY (shifter_id, validity_end, transaction_end, mode)
) PARTITION BY LIST (mode);

SELECT create_bitemporal_partitions('shifter', ARRAY['shifter_id']);

CREATE TABLE shifter_localization (
    shifter_id                  BIGINT       NOT NULL
  , validity_start              timestamptz  NOT NULL DEFAULT transaction_timestamp()
  , validity_end                timestamptz  NOT NULL DEFAULT '3000-01-01 00:00:00+00'::timestamptz CHECK (validity_start < validity_end)
  , transaction_start           timestamptz  NOT NULL DEFAULT transaction_timestamp()
  , transaction_end             timestamptz  NOT NULL DEFAULT '3000-01-01 00:00:00+00'::timestamptz
  , mode                        bitemporal_mode NOT NULL DEFAULT 'active'
  , language                    CITEXT       NOT NULL CHECK (length(language) BETWEEN 2 AND 3)
  , country                     CITEXT       NOT NULL CHECK (length(country) = 0 OR length(country) BETWEEN 2 AND 3)
  , display_name                TEXT         NOT NULL CHECK (length(display_name) BETWEEN 6 AND 64)
  , tooltip                     TEXT         NOT NULL CHECK (length(tooltip) BETWEEN 15 AND 256)
  , description                 TEXT         NOT NULL CHECK (length(description) > 25)
  , insertion_metadata          JSONB        NOT NULL
  , CONSTRAINT shifter_localization_physical_pk PRIMARY KEY (shifter_id, language, country, validity_end, transaction_end, mode)
) PARTITION BY LIST (mode);

SELECT create_bitemporal_partitions('shifter_localization', ARRAY['shifter_id', 'language', 'country']);

CREATE TABLE shifter_input (
    shifter_id              BIGINT       NOT NULL
  , metric_id                   BIGINT       NOT NULL
  , validity_start              timestamptz  NOT NULL DEFAULT transaction_timestamp()
  , validity_end                timestamptz  NOT NULL DEFAULT '3000-01-01 00:00:00+00'::timestamptz CHECK (validity_start < validity_end)
  , transaction_start           timestamptz  NOT NULL DEFAULT transaction_timestamp()
  , transaction_end             timestamptz  NOT NULL DEFAULT '3000-01-01 00:00:00+00'::timestamptz
  , mode                        bitemporal_mode NOT NULL DEFAULT 'active'
  , input_order                 SMALLINT     NOT NULL
  , CONSTRAINT shifter_input_physical_pk   PRIMARY KEY (shifter_id, metric_id, validity_end, transaction_end, mode)
) PARTITION BY LIST (mode);

SELECT create_bitemporal_partitions('shifter_input', ARRAY['shifter_id', 'metric_id']);

CREATE TABLE shifter_output (
    shifter_id              BIGINT       NOT NULL
  , metric_id                   BIGINT       NOT NULL
  , validity_start              timestamptz  NOT NULL DEFAULT transaction_timestamp()
  , validity_end                timestamptz  NOT NULL DEFAULT '3000-01-01 00:00:00+00'::timestamptz CHECK (validity_start < validity_end)
  , transaction_start           timestamptz  NOT NULL DEFAULT transaction_timestamp()
  , transaction_end             timestamptz  NOT NULL DEFAULT '3000-01-01 00:00:00+00'::timestamptz
  , mode                        bitemporal_mode NOT NULL DEFAULT 'active'
  , CONSTRAINT shifter_output_physical_pk   PRIMARY KEY (shifter_id, metric_id, validity_end, transaction_end, mode)
) PARTITION BY LIST (mode);

SELECT create_bitemporal_partitions('shifter_output', ARRAY['shifter_id', 'metric_id']);

-- ============================================================================
-- SECTION 5: TIMESERIES
-- ============================================================================

CREATE TABLE ts_doc_type (
    ts_doc_type_id              BIGINT       GENERATED ALWAYS AS IDENTITY
  , ns_id                       BIGINT       NOT NULL
  , validity_start              timestamptz  NOT NULL DEFAULT transaction_timestamp()
  , validity_end                timestamptz  NOT NULL DEFAULT '3000-01-01 00:00:00+00'::timestamptz CHECK (validity_start < validity_end)
  , transaction_start           timestamptz  NOT NULL DEFAULT transaction_timestamp()
  , transaction_end             timestamptz  NOT NULL DEFAULT '3000-01-01 00:00:00+00'::timestamptz
  , mode                        bitemporal_mode NOT NULL DEFAULT 'active'
  , key                         citext       NOT NULL CHECK (LENGTH(key) BETWEEN 6 AND 64)
  , insertion_metadata          JSONB        NOT NULL
  , CONSTRAINT doc_type_physical_pk          PRIMARY KEY (ts_doc_type_id, validity_end, transaction_end, mode)
) PARTITION BY LIST (mode);

SELECT create_bitemporal_partitions('ts_doc_type', ARRAY['ts_doc_type_id']);

CREATE TABLE ts_doc_type_localization (
    ts_doc_type_id              BIGINT       NOT NULL
  , validity_start              timestamptz  NOT NULL DEFAULT transaction_timestamp()
  , validity_end                timestamptz  NOT NULL DEFAULT '3000-01-01 00:00:00+00'::timestamptz CHECK (validity_start < validity_end)
  , transaction_start           timestamptz  NOT NULL DEFAULT transaction_timestamp()
  , transaction_end             timestamptz  NOT NULL DEFAULT '3000-01-01 00:00:00+00'::timestamptz
  , mode                        bitemporal_mode NOT NULL DEFAULT 'active'
  , language                    CITEXT       NOT NULL CHECK (length(language) BETWEEN 2 AND 3)
  , country                     CITEXT       NOT NULL CHECK (length(country) = 0 OR length(country) BETWEEN 2 AND 3)
  , display_name                TEXT         NOT NULL CHECK (length(display_name) BETWEEN 6 AND 64)
  , tooltip                     TEXT         NOT NULL CHECK (length(tooltip) BETWEEN 15 AND 256)
  , description                 TEXT         NOT NULL CHECK (length(description) > 25)
  , insertion_metadata          JSONB        NOT NULL
  , CONSTRAINT ts_doc_type_localization_physical_pk PRIMARY KEY (ts_doc_type_id, language, country, validity_end, transaction_end, mode)
) PARTITION BY LIST (mode);

SELECT create_bitemporal_partitions('ts_doc_type_localization', ARRAY['ts_doc_type_id', 'language', 'country']);

CREATE TABLE ts_doc_row_type (
    ts_doc_row_type_id          BIGINT       GENERATED ALWAYS AS IDENTITY
  , ts_doc_type_id              BIGINT       NOT NULL
  , metric_id                   BIGINT       NOT NULL
  , validity_start              timestamptz  NOT NULL DEFAULT transaction_timestamp()
  , validity_end                timestamptz  NOT NULL DEFAULT '3000-01-01 00:00:00+00'::timestamptz CHECK (validity_start < validity_end)
  , transaction_start           timestamptz  NOT NULL DEFAULT transaction_timestamp()
  , transaction_end             timestamptz  NOT NULL DEFAULT '3000-01-01 00:00:00+00'::timestamptz
  , mode                        bitemporal_mode NOT NULL DEFAULT 'active'
  , key                         citext       NOT NULL CHECK (LENGTH(key) BETWEEN 6 AND 64)
  , insertion_metadata          JSONB        NOT NULL
  , CONSTRAINT doc_row_type_physical_pk      PRIMARY KEY (ts_doc_row_type_id, validity_end, transaction_end, mode)
) PARTITION BY LIST (mode);

SELECT create_bitemporal_partitions('ts_doc_row_type', ARRAY['ts_doc_row_type_id']);

CREATE TABLE ts_doc_row_type_localization (
    ts_doc_row_type_id          BIGINT       NOT NULL
  , validity_start              timestamptz  NOT NULL DEFAULT transaction_timestamp()
  , validity_end                timestamptz  NOT NULL DEFAULT '3000-01-01 00:00:00+00'::timestamptz CHECK (validity_start < validity_end)
  , transaction_start           timestamptz  NOT NULL DEFAULT transaction_timestamp()
  , transaction_end             timestamptz  NOT NULL DEFAULT '3000-01-01 00:00:00+00'::timestamptz
  , mode                        bitemporal_mode NOT NULL DEFAULT 'active'
  , language                    CITEXT       NOT NULL CHECK (length(language) BETWEEN 2 AND 3)
  , country                     CITEXT       NOT NULL CHECK (length(country) = 0 OR length(country) BETWEEN 2 AND 3)
  , display_name                TEXT         NOT NULL CHECK (length(display_name) BETWEEN 6 AND 64)
  , tooltip                     TEXT         NOT NULL CHECK (length(tooltip) BETWEEN 15 AND 256)
  , description                 TEXT         NOT NULL CHECK (length(description) > 25)
  , insertion_metadata          JSONB        NOT NULL
  , CONSTRAINT ts_doc_row_type_localization_physical_pk PRIMARY KEY (ts_doc_row_type_id, language, country, validity_end, transaction_end, mode)
) PARTITION BY LIST (mode);

SELECT create_bitemporal_partitions('ts_doc_row_type_localization', ARRAY['ts_doc_row_type_id', 'language', 'country']);

CREATE TABLE ts_doc (
    ts_doc_id                   BIGINT       GENERATED ALWAYS AS IDENTITY
  , ts_doc_type_id              BIGINT       NOT NULL
  , as_of                       TIMESTAMPTZ  NOT NULL DEFAULT transaction_timestamp()
  , key                         CITEXT       NOT NULL CHECK (length(key) BETWEEN 6 AND 64)
  , storage_config_id           INTEGER      NOT NULL
  , location                    TEXT         NOT NULL
  , insertion_metadata          JSONB        NOT NULL
  , CONSTRAINT ts_doc_physical_pk            PRIMARY KEY (ts_doc_id)
)
;

CREATE TABLE ts_doc_localization (
    ts_doc_id                   BIGINT       NOT NULL
  , validity_start              timestamptz  NOT NULL DEFAULT transaction_timestamp()
  , validity_end                timestamptz  NOT NULL DEFAULT '3000-01-01 00:00:00+00'::timestamptz CHECK (validity_start < validity_end)
  , transaction_start           timestamptz  NOT NULL DEFAULT transaction_timestamp()
  , transaction_end             timestamptz  NOT NULL DEFAULT '3000-01-01 00:00:00+00'::timestamptz
  , mode                        bitemporal_mode NOT NULL DEFAULT 'active'
  , language                    CITEXT       NOT NULL CHECK (length(language) BETWEEN 2 AND 3)
  , country                     CITEXT       NOT NULL CHECK (length(country) = 0 OR length(country) BETWEEN 2 AND 3)
  , display_name                TEXT         NOT NULL CHECK (length(display_name) BETWEEN 6 AND 64)
  , tooltip                     TEXT         NOT NULL CHECK (length(tooltip) BETWEEN 15 AND 256)
  , description                 TEXT         NOT NULL CHECK (length(description) > 25)
  , insertion_metadata          JSONB        NOT NULL
  , CONSTRAINT ts_doc_localization_physical_pk PRIMARY KEY (ts_doc_id, language, country, validity_end, transaction_end, mode)
) PARTITION BY LIST (mode);

SELECT create_bitemporal_partitions('ts_doc_localization', ARRAY['ts_doc_id', 'language', 'country']);

CREATE TABLE ts_doc_tags (
    ts_doc_id                   BIGINT       NOT NULL REFERENCES ts_doc
  , tag_id                      BIGINT       NOT NULL
  , tag_index                   SMALLINT     NOT NULL
  , CONSTRAINT ts_doc_tag_pk    PRIMARY KEY (ts_doc_id, tag_id)
  , CONSTRAINT ts_doc_tag_ak    UNIQUE (ts_doc_id, tag_index)
)
;

CREATE TABLE ts_doc_input_lineage (
    ts_doc_id                   BIGINT       NOT NULL REFERENCES ts_doc
  , input_id                    BIGINT       NOT NULL REFERENCES ts_doc
  , input_index                 SMALLINT     NOT NULL
  , CONSTRAINT ts_doc_lineage_pk    PRIMARY KEY (ts_doc_id, input_id)
  , CONSTRAINT ts_doc_lineage_ak    UNIQUE (ts_doc_id, input_index)
)
;
