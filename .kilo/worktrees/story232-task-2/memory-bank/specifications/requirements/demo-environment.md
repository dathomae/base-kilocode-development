# Demonstration Environment Requirements

This document specifies requirements for the ChronoCone demonstration environment, designed to run on a demonstrator's laptop for presentation and evaluation purposes with limited-sized plans.

## Runtime Environment

Demonstrations will be made from the containerized development environment, which consists of two docker containers, one for the services and one for the database.

Being able to easily start or shutdown/restart an environment during demonstrations is important and should be made as frictionless as possible. Convenience PDD scripts are provided to handle the lifecycle of the demo environment, including initialization, startup, shutdown, and restart when used from within an agentic IDE. Shell scripts are provided for use from the command line for those purposes for use either inside or outside of an agentic IDE.

## Core Server

The Core Server serves as the primary entry point and CLI for the demonstration environment. Its responsibilities include:

- **System Entry Point**: Acting as the main interface for starting the system in a local laptop or demonstration environment.
- **Service Orchestration**: Responsible for starting and managing other dependent services required for the demo.
- **Automated Deployment**: Configured to automatically start within a Docker container, orchestrated by a local Kubernetes environment (using Kind).
- **Service Discovery**: Providing an endpoint for other services to discover the locations of running components.

## External API Simulator

_Not Yet Written_
