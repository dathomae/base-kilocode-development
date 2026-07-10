# ChronoCone

A time-series data processing and planning system designed to solve planning problems across different domains including logistics, manufacturing, and healthcare.

## Overview

ChronoCone provides a flexible framework supporting automated planning, human-in-the-loop planning, and collaborative planning across different organizational levels. The system manages plan templates and instances with full lifecycle management, metrics collection, and external system integration.

## Workspace Structure

This project uses a VSCode multi-root workspace configuration with three roots:

```
CHRONOCONE/                    # Top-level project root
├── .vscode/
│   └── chronocone.code-workspace   # Multi-root workspace configuration
├── src/
│   ├── backend/                    # Go/Huma backend API
│   │   ├── handlers/               # HTTP route handlers
│   │   ├── store/                  # In-memory store (DB layer pending)
│   │   ├── main.go                 # Entry point
│   │   └── go.mod                  # Go module definition
│   ├── chronocone_planning_language/ # CPL compiler (Go + ANTLR)
│   └── frontend/                   # React/TypeScript frontend
│       ├── node_modules/           # npm packages (not in git)
│       ├── src/
│       ├── package.json            # npm dependencies
│       └── ...
├── memory-bank/                    # Project documentation and planning
└── ...
```

### Workspace Roots

1. **Top-level (`CHRONOCONE/`)** - Project root containing workspace configuration, documentation, and shared resources
2. **Backend (`src/backend/`)** - Go Huma backend API
3. **Frontend (`src/frontend/`)** - React/TypeScript frontend application

## Quick Start (DevContainers)

The recommended and easiest way to develop ChronoCone is using a **VSCode DevContainer**. This containerises all dependencies (Go, Node.js) so you don't need them installed locally.

### Prerequisites

- Docker Desktop or Docker Engine
- VSCode with the "Dev Containers" extension

### Opening the Workspace

1. Open the project folder in VSCode.
2. When prompted, click **Reopen in Container** (or press F1 and select `Dev Containers: Reopen in Container`).
3. The container will build and automatically install backend and frontend dependencies.

*(For detailed manual setup instructions without Docker, see [DEVELOPMENT.md](DEVELOPMENT.md))*

### Backend Setup

1. Navigate to the backend directory:

   ```bash
   cd src/backend
   ```

2. Run the development server:

   ```bash
   go run main.go
   ```

   The backend API will be available at `http://localhost:8080`
   Interactive API docs are available at `http://localhost:8080/docs`

### Frontend Setup

1. Navigate to the frontend directory:

   ```bash
   cd src/frontend
   ```

2. Install dependencies:

   ```bash
   npm install
   ```

3. Run the development server:

   ```bash
   npm run dev
   ```

   The frontend will be available at `http://localhost:5173` (or another port if 5173 is in use)

## Development Workflow

### Running Both Services

For full-stack development, run both services concurrently:

1. Open two terminal instances in VSCode
2. In the backend terminal: `cd src/backend && go run main.go`
3. In the frontend terminal: `cd src/frontend && npm run dev`

### CI/CD

The project includes GitHub Actions workflows for continuous integration:

- `.github/workflows/ci.yml` - Builds and tests Go backend, CPL parser, and frontend

## Project Documentation

Detailed documentation is located in the `memory-bank/` directory:

- [`memory-bank/brief.md`](memory-bank/brief.md) - Project overview and purpose
- [`memory-bank/design.md`](memory-bank/design.md) - System design documentation
- [`memory-bank/requirements.md`](memory-bank/requirements.md) - Requirements specification
- [`memory-bank/stories/`](memory-bank/stories/) - Development stories and tasks

## Technology Stack

- **Backend**: Go 1.22+, Huma v2, chi
- **Planning Language**: Go 1.22+, ANTLR 4
- **Frontend**: React 18+, TypeScript, Vite
- **Database**: PostgreSQL (pending; currently in-memory store)
- **Testing**: Go test (backend), Vitest (frontend)
- **CI/CD**: GitHub Actions

## License

Copyright (c) 2026 Kevin Finley and Doug Thomae
