# Story028: Demo Environment Core and Server Requirements Exploration

## Goal

Analyze and define the major requirements for the ChronoCone demonstration environment's **core server and general infrastructure**, designed to run on a demonstrator's laptop for presentation and evaluation purposes with limited-sized plans. This story captures the specification for the demo environment's foundational components including the core server, deployment procedures, and production differences.

**Scope Note:** This story focuses on the core demo infrastructure. External API simulator and sample data requirements are covered separately in [Story035](Story035_external-api-simulator-and-sample-data-requirements.md).

At the discretion of the person working on this story, architecture or design decisions may be made and recorded in the architecture or design documents. Additionally, one or more implementation stories may be defined to capture work that extends beyond the requirements definition itself. The exact result for this story will depend on the complexity of the requirements discovered and the amount of discussion needed to reach clarity.

## References

- [ChronoCone Concepts](../specifications/chronocone-concepts.md) - Key concepts
- [Introduction](../specifications/requirements/introduction.md) - System overview and deployment contexts
- [Demo Environment Requirements](../specifications/requirements/demo-environment.md) - Target document to populate
- All other requirements documents - The demo environment implements a subset of general requirements
- Existing Stories: Story008 (Docker demo container), Story023 (containerized development)

## Dependencies

None - This story has no precedents. As the root story for the demo environment core server, this work must be specified and built first. Other demo environment components (such as the external API simulator and sample data in Story035) will depend on this story.

## Dependent Stories

- [Story035](Story035_external-api-simulator-and-sample-data-requirements.md) - External API Simulator and Sample Data requirements (elaborates simulator and sample data aspects)
- Implementation stories will be created after demo environment requirements are defined

## Context

From [introduction.md](../specifications/requirements/introduction.md):

**Demonstration Environment:**

- Designed for use on a demonstrator's laptop
- Supports limited-sized plans for presentation and evaluation
- Self-contained, easy-to-set-up instance
- Showcases core functionality without extensive infrastructure

**Demo Environment Scope:**
The demo-environment.md file covers:

1. **Core Server** - Covered by this story (Story028)
2. **Setup and Deployment** - Covered by this story (Story028)
3. **Production Differences** - Covered by this story (Story028)
4. **External API Simulator** - Elaborated in Story035
5. **Sample Data and Plans** - Elaborated in Story035

**Existing Demo Work:**

- Story008 created a Docker demo container concept
- Story023 established containerized development environment
- Demo environment builds on these foundations

## Purpose of This Section

The demo-environment.md requirements document's **core server and infrastructure sections** must specify:

1. Demo environment scope and limitations
2. Core server requirements for demo
3. Setup and deployment procedures
4. Differences from production environments

**Note:** External API simulator and sample data requirements are elaborated in [Story035](Story035_external-api-simulator-and-sample-data-requirements.md).

## Major Architectural Elements to Elaborate

### 1. Demo Environment Scope

- Plan size limitations (data volume, metric count)
- Supported features vs production
- Performance expectations (appropriate for laptop deployment)
- Persistence model (ephemeral vs saved state)
- Reset and refresh capabilities

### 2. Core Server Requirements

- Services to include in demo (Metrics, Ontology, etc.)
- Simplified service architecture (single process vs microservices)
- In-memory vs persistent storage tradeoffs
- Calculation engine capabilities and limitations
- UI availability and features

### 3. Setup and Deployment

- Docker-based deployment (building on Story008)
- Single-command startup
- Configuration options
- Port mappings and access
- Documentation and tutorials

### 4. Production Differences

- What's simplified or omitted
- Performance expectations
- Data persistence warnings
- Migration path to production (if any)
- Feature comparison matrix

## Exploration Tasks

1. **Define demo environment scope** - Completed
   - Specify plan size limitations
   - Define supported features
   - Document performance expectations

2. **Specify core server requirements** - Completed
   - Define services to include
   - Specify architecture simplifications
   - Document storage tradeoffs

3. **Specify setup and deployment** - Completed
   - Define Docker configuration
   - Specify startup procedures
   - Document access patterns

4. **Define production differences** - Completed
   - Specify simplifications
   - Define migration paths
   - Document feature gaps

**Note:** Tasks for external API simulator and sample data are defined in [Story035](Story035_external-api-simulator-and-sample-data-requirements.md).

## Next Steps (To Be Determined)

After completing this exploration, determine:

- Can the demo environment core requirements be written as a single implementation story?
- What dependencies exist between core demo environment and Story035 (simulator/sample data)?
- Should this story and Story035 be implemented together or sequentially?

## Notes

- Demo environment must be truly easy to set up - single command ideal
- Performance on typical laptops (8GB RAM, integrated graphics) must be acceptable
- Consider video tutorials or guided tours for first-time users
- Demo environment should showcase the "ChronoCone difference" - what makes it special
