# Epic Restraining Order (ERO) Template

> **Note:** Pre-2026-09-16 EROs are frozen in `memory-bank/specifications/design/historical-archive/`
> (paths preserved relative to `specifications/design/`). This template is for new EROs, which are
> created under the live ERO directory for the current organization. Epics are also archived, so the
> "Related Epics" links in this template refer to the pre-reorganization layout.

An Epic Restraining Order (ERO) halts architectural chaos by documenting a single technical decision and its context, consequences, and alternatives. EROs capture the reasoning behind important technical choices to preserve institutional knowledge and enforce order on future decisions.

## When to Issue an ERO

Issue an ERO when making significant technical decisions related to managing confusion from incremental implementation, including:

- Choosing between competing technologies or approaches during phased implementation
- Introducing temporary limitations during incremental development
- Establishing patterns or conventions that affect multiple components during transition
- Making trade-offs between competing quality attributes (performance vs. maintainability, etc.)
- Documenting the gap between current implementation state and final architecture

## What Belongs in EROs

EROs capture design decisions specifically related to managing confusion from incremental implementation:

- **Current vs Final State Gaps**: Documentation of what exists now versus what will exist when implementation is complete
- **Technical Decisions for Incomplete Implementations**: Design choices made specifically to handle transitional states
- **Clarifications for Mismatches**: Explanations needed due to differences between current and final system state
- **Key Data Structures**: Data structures that bridge current and final states
- **Transitional Algorithms**: Algorithms for managing implementation gaps

## What Does NOT Belong in EROs

EROs are not for general architecture or design decisions:

- **General Critical Architecture Decisions** → Belong in architecture documentation
- **General Design Decisions** → Belong in design documentation
- **Requirements** → Belong in requirements documents

Only decisions specifically related to incremental implementation confusion belong in EROs.

## How EROs Differ from Other Documents

**vs Architecture/Design Docs**: Architecture and design documents describe the system as it ultimately will be. EROs describe design decisions made specifically to manage confusion from the gap between current and final state.

**vs Epics and Stories**: Epics and stories are about tracking, planning, and coordinating work. EROs capture design decisions related to incremental implementation confusion, including key data structures and algorithms used to bridge implementation gaps.

## File Naming Convention

ERO files use the format `ERONNN_<short-decision-name>.md`, where `NNN` is a unique three-digit number (e.g., `ERO001_database-technology-choice.md`).

## Table of Contents

Each ERO appears in the relevant `toc.md` file (typically in the architecture or design directory) with its current decision status and a brief summary.

Valid decision statuses are:

- **Draft**: The decision is still being considered and may change
- **Decided**: The decision has been made and is in effect
- **Superseded**: The decision has been replaced by a newer ERO (include link to the superseding ERO)

## Header Section

### 1. Title

Provide a clear, concise title that describes the decision being documented.

### 2. Purpose/Goal

State the purpose or goal of this ERO. What technical decision needs to be made and why? This should be a brief statement (1-3 sentences) that captures the essence of what this ERO will document.

### 3. Implementation Status

Indicate the current state of implementation:

- **Not Yet Implemented** - The decision has been made but not yet executed
- **Implemented** - The decision has been fully executed

If implementation is associated with specific epics, list them:

**Related Epics:**

- [EpicNNN\_<short-purpose>.md](../epics/EpicNNN_<short-purpose>.md) - <status>

If no epics are associated, state "None - This ERO documents a decision not tied to specific work items."

## Body Section

### 1. Context

Describe the situation that necessitates this decision. Include:

**Situation:** What is the problem or opportunity being addressed? What is the current state?

**Constraints:** What limitations or boundaries must be respected? These may be technical, organizational, budgetary, or time-based.

**Requirements:** What functional or non-functional requirements drive this decision? Reference specific requirements documents where applicable.

**Assumptions:** What assumptions are being made about the problem space, technology, or future conditions? Be explicit about what is assumed to be true.

### 2. Evaluation Criteria / Strategy Fit

Describe the criteria used to evaluate options, or explain how this decision fits into a broader architectural strategy.

**Evaluation Criteria:**

- Criterion 1: Description and weight/importance
- Criterion 2: Description and weight/importance
- Criterion 3: Description and weight/importance

**Strategic Context:**

If this ERO implements a broader strategy, reference that strategy here:

- Link to strategic document or pattern
- Explanation of how this decision aligns with or advances the strategy

### 3. Options Considered

List and describe the options that were evaluated. Limit this to no more than three options. For each option, provide sufficient technical detail to understand exactly what is being proposed.

#### Option 1: <Name/Description>

**Description:**
Detailed technical description of this option.

**Pros:**

- Advantage 1
- Advantage 2

**Cons:**

- Disadvantage 1
- Disadvantage 2

#### Option 2: <Name/Description>

**Description:**
Detailed technical description of this option.

**Pros:**

- Advantage 1
- Advantage 2

**Cons:**

- Disadvantage 1
- Disadvantage 2

#### Option 3: <Name/Description> (if applicable)

**Description:**
Detailed technical description of this option.

**Pros:**

- Advantage 1
- Advantage 2

**Cons:**

- Disadvantage 1
- Disadvantage 2

### 4. Decision

State the option chosen and provide a clear rationale.

**Chosen Option:** Option N - <Name/Description>

**Rationale:**
Explain why this option was selected over the alternatives. Reference the evaluation criteria and explain how the chosen option best satisfies them. Be explicit about the trade-offs being accepted.

### 5. Implications

Describe the consequences of this decision. Consider both positive and negative implications, and implications for different stakeholders or system aspects.

**Positive Implications:**

- Benefit 1
- Benefit 2

**Negative Implications / Trade-offs:**

- Cost or limitation 1
- Cost or limitation 2

**Impact on System Qualities:**

| Quality Attribute | Impact | Notes             |
| ----------------- | ------ | ----------------- |
| Performance       | +/-/=  | Brief explanation |
| Scalability       | +/-/=  | Brief explanation |
| Maintainability   | +/-/=  | Brief explanation |
| Security          | +/-/=  | Brief explanation |

**Future Considerations:**

- What might change this decision in the future?
- Under what conditions should this ERO be revisited?
- Are there follow-up decisions that need to be made?

## Footer Section (Optional)

Include any additional information relevant to the ERO:

1. **Open Questions**: Questions that need answers before the decision can be finalized
2. **References**: Links to related documents, research, or external resources
3. **Notes**: Any other relevant information or background
4. **Decision Log**: If the decision changed over time, briefly document the evolution

## Example ERO

Below is a minimal example showing all sections properly filled out:

```markdown
# ERO001: Database Technology Selection

## Title

Database Technology Selection for Primary Data Store

## Purpose/Goal

Select a primary relational database technology for storing configuration data, user information, and transactional records.

## Decision Status

Decided

## Implementation Status

Implemented

**Related Epics:**

- None - This decision predates the epic structure.

## Context

### Situation

The system requires a relational database for storing configuration data, user information, and transactional records. We need to select a primary database technology that will serve as the main data store.

### Constraints

- Must be open-source or have reasonable licensing costs
- Team has limited DBA resources; needs low operational overhead
- Must support ACID transactions for data consistency
- Must have strong Python and Go driver support

### Requirements

- Support for complex queries with joins across multiple tables
- JSON support for semi-structured data
- Replication capabilities for high availability
- Migration tooling available

### Assumptions

- Data volume will not exceed 100GB in the first two years
- The team will grow to include DBA expertise within 18 months
- Cloud deployment on AWS or GCP is likely

## Evaluation Criteria

1. **Operational Complexity** (High importance): How much expertise is required to run in production?
2. **Ecosystem Maturity** (Medium importance): Availability of tools, libraries, and community support
3. **Performance Characteristics** (Medium importance): Query performance for expected access patterns
4. **Cost** (Medium importance): Licensing and operational costs

## Options Considered

### Option 1: PostgreSQL

**Description:**
Use PostgreSQL as the primary relational database. PostgreSQL is an open-source, object-relational database with strong SQL compliance and extensive feature set.

**Pros:**

- Excellent SQL standard compliance
- Strong JSON support via JSONB
- Mature ecosystem with many tools
- Excellent Python (psycopg2) and Go (pgx) driver support
- Built-in replication

**Cons:**

- Can be complex to tune for high performance
- Single-master replication limits write scaling

### Option 2: MySQL / MariaDB

**Description:**
Use MySQL or MariaDB as the primary relational database. These are widely-used open-source relational databases.

**Pros:**

- Very widely adopted with extensive tooling
- Easier to find operational expertise
- Good replication support

**Cons:**

- Less strict SQL compliance than PostgreSQL
- JSON support less mature than PostgreSQL
- Oracle licensing concerns for MySQL

### Option 3: SQLite

**Description:**
Use SQLite for embedded, serverless database storage.

**Pros:**

- Zero configuration
- Single file database
- Excellent for testing and development

**Cons:**

- Not suitable for concurrent write-heavy workloads
- Limited replication and high availability options
- Not appropriate for multi-user applications

## Decision

**Chosen Option:** Option 1 - PostgreSQL

**Rationale:**
PostgreSQL best satisfies our evaluation criteria. It offers the lowest operational complexity for our use case given the team's existing familiarity. Its superior JSON support (JSONB) aligns with our semi-structured data requirements. The ecosystem maturity ensures we can find solutions to problems and hire expertise. While write scaling is limited to single-master, our data volume assumptions indicate this will not be a constraint for at least two years.

The decision to not use SQLite was straightforward given our multi-user, concurrent access requirements. MySQL/MariaDB was a close second, but PostgreSQL's stricter standards compliance and better JSON support tipped the balance.

## Implications

### Positive Implications

- Strong foundation for data consistency and integrity
- Rich query capabilities for complex reporting requirements
- Good migration path to managed services (AWS RDS, Cloud SQL)
- Excellent tooling for schema management and migrations

### Negative Implications / Trade-offs

- Write scalability will eventually require application-level sharding or migration to a distributed database
- Some operational complexity for backups, replication setup, and performance tuning
- Team needs to develop PostgreSQL-specific expertise for optimization

### Impact on System Qualities

| Quality Attribute | Impact | Notes                                            |
| ----------------- | ------ | ------------------------------------------------ |
| Performance       | +      | Good read performance; single-master write limit |
| Scalability       | =      | Vertical scaling primarily; horizontal for reads |
| Maintainability   | +      | Clear schema, strong constraints, good tooling   |
| Security          | +      | Mature security model, good access control       |

### Future Considerations

- This decision should be revisited if data volume exceeds 100GB or write throughput requirements increase significantly
- Consider read replicas for scaling read-heavy workloads
- Evaluate connection pooling (PgBouncer) if connection count becomes an issue

## Notes

- Decision made after team discussion on 2024-01-22
- PostgreSQL 15 selected as the target version
- pgAdmin considered but not adopted; prefer psql and migration scripts
```
