# Timecone.ai and ChronoCone Requirements

## Part 1: System Overview

### What is Timecone.ai?

Timecone.ai is a visionary planning system designed to act as an "Institutional Brain." It moves beyond "flat time" to manage multiple possible futures simultaneously, providing a flexible framework that supports automated planning, human-in-the-loop planning, and collaborative planning across different organizational levels.

The system overcomes the "Consensus Trap"—the organizational tendency to suppress discordance to force a singular, often inaccurate, future plan. Timecone.ai embraces discordance by allowing Permissionless Assertion Layers, where different domains can model conflicting futures.

At the core of Timecone.ai is the **ChronoCone** calculation engine. ChronoCone provides the deterministic, rule-based calculating substrate needed because pure agentic AI currently fails at complex planning (the "Hallucination Kills Process" problem). By grounding agents in rigid math, ChronoCone allows advanced agentic AI to reason safely at macro levels.

The name "ChronoCone" represents widening "cones" of possible future states. Plans can expand either through refinement, where general plans develop into specific ones (for example, country-level plans break into regional plans), or through temporal extension, where short-term plans extend into longer horizons.

The system serves diverse planning domains including logistics, manufacturing, and healthcare. These domains share common planning challenges such as coordinating across organizational boundaries, managing uncertainty, and balancing resource constraints against operational goals.

## Part 2: Deployment Environments

ChronoCone is designed to operate across multiple deployment environments, each tailored to specific use cases and constraints. The requirements are organized around these deployment contexts.

### Demonstration Environment

Designed for use on a demonstrator's laptop. It supports limited-sized plans for presentation and evaluation purposes. This environment provides a self-contained, easy-to-set-up instance of ChronoCone that showcases core functionality without requiring extensive infrastructure. Requirements for the demonstration environment specify components such as the core server and external API simulator.

### On-Premise Environment

Operates at a customer facility, providing full-scale planning capabilities with enterprise-grade reliability, security, and performance. This environment is designed for real-world operational use with integration into existing organizational systems and processes.

### Cloud Environment

Provides ChronoCone as a hosted service, offering scalability and accessibility without requiring on-premises infrastructure. This deployment model enables organizations to leverage ChronoCone capabilities without managing their own server infrastructure.

**Note:** Current requirements focus primarily on the demonstration environment. On-premise and cloud environment requirements will be developed in future iterations.

## Part 3: Documentation Organization

The requirements documentation is organized into two main categories:

### General Requirements

General requirements apply to all deployment environments. These requirements specify core system functionality that remains consistent regardless of where or how ChronoCone is deployed.

| Topic | File |
| ----- | ---- |
| Metrics and Calculation Structure | [`metrics-and-calculation.md`](metrics-and-calculation.md) |
| Web UI | [`web-ui.md`](web-ui.md) |
| Plugin UI and Functionality | [`plugin-ui-and-functionality.md`](plugin-ui-and-functionality.md) |
| Agents and Agentic Interaction | [`agents-and-agentic-interaction.md`](agents-and-agentic-interaction.md) |
| MCP Servers | [`mcp-servers.md`](mcp-servers.md) |
| Versioning | [`versioning.md`](versioning.md) |
| Data and Plan Definition Lineage | [`data-and-plan-definition-lineage.md`](data-and-plan-definition-lineage.md) |
| User Identity, Authentication, and Authorization | [`identity-authentication-authorization.md`](identity-authentication-authorization.md) |
| Metadata | [`metadata.md`](metadata.md) |
| Data Stores | [`data-stores.md`](data-stores.md) |
| Connectors and Drivers | [`connectors-and-drivers.md`](connectors-and-drivers.md) |

### Environment-Specific Requirements

Environment-specific requirements address the unique needs of particular deployment contexts. These requirements specify components, configurations, and constraints that apply only to specific environments.

| Topic | File |
| ----- | ---- |
| Demonstration Environment | [`demo-environment.md`](demo-environment.md) |

The first implementation focuses on the demonstration environment requirements, with support from the general requirements.
