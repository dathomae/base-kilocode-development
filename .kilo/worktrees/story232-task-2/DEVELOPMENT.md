# Chronocone Development Setup Guide

This guide provides step-by-step instructions for setting up a local development environment for the Chronocone project.

## Prerequisites

The absolute easiest way to develop Chronocone is via **VSCode DevContainers**. Ensure you have the following installed:

1. **Docker** - (Docker Desktop or Docker Engine)
2. **Visual Studio Code**
3. **Dev Containers Extension** - (`ms-vscode-remote.remote-containers`)

There is **no need** to install Go, Node.js, or PostgreSQL locally.

## Getting Started

1. Open the project root in VSCode.
2. When prompted "Folder contains a Dev Container configuration file", click **Reopen in Container** (or press F1 and run `Dev Containers: Reopen in Container`).
3. VSCode will build the Docker container and attach your environment to it.
4. All dependencies (`go mod download` for Go and `pnpm` for Node.js) will be **automatically installed** during container creation.

## Backend Setup (Inside Container)

```bash
cd src/backend && go mod download
```

*(Note: If you use the DevContainer workflow, this is handled automatically!)*

## Start Backend Server

```bash
cd src/backend && go run main.go
```

The backend API will be available at `http://localhost:8080` and the interactive API docs at `http://localhost:8080/docs`.

## Frontend Setup

The Chronocone frontend is a React application built with Vite.

To install the frontend dependencies (if not using the automated DevContainer setup):

```bash
cd src/frontend && pnpm install
```

## Start Frontend Server

```bash
cd src/frontend && pnpm run dev
```

The application will be available at `http://localhost:5173`.

## PDD Scripts for Agentic IDEs

If you are running this project in an agentic IDE, you can use the Prompt-Driven Development (PDD) scripts located in the `scripts/` directory.
