# Story023: Fully Containerized Development Environment

## Goal

Transition the ChronoCone project to a fully containerized development environment using `.devcontainer`. This will allow developers to open the project in VSCode or GitHub Codespaces and have a complete, ready-to-use environment without installing local tools. The environment must support both the Python backend and the React/Node.js frontend, as well as seamlessly handle the PostgreSQL database requirements.

## References

- [DEVELOPMENT.md](../../DEVELOPMENT.md)
- [README.md](../../README.md)
- [Workspace Setup](Story001_workspace-setup.md)

## Dependencies

- None

## Tasks

1. Enhance `.devcontainer/Dockerfile` for Full-Stack Development - Done
   a. Change base image to `ubuntu:24.04`. - Done
   b. Install Python 3.12 securely along with necessary system dependencies (git, curl, build-essential) via `apt-get`. - Done
   c. Install Node.js (v20 LTS recommended) and `pnpm` to support the frontend workspace securely and efficiently. - Done
   d. Install `uv`, an extremely fast Python package manager written in Rust, to replace `pip`. - Done
   e. Remove the `COPY requirements.txt ./` and `pip install` from the Dockerfile. Because the dependencies exist in nested directories, bootstrapping will be moved to the `devcontainer.json` lifecycle hooks to ensure volume mounts are fully available. - Done
   f. Remove the hardcoded `CMD ["python", "main.py"]` as it blocks normal devcontainer operation. - Done

2. Update `.devcontainer/docker-compose.yml` - Done
   a. Configure the `app` service to build from the updated Dockerfile. - Done
   b. Add a `db` service for PostgreSQL using `postgres:15` to run alongside the app. - Done
   c. Configure a persistent volume for the PostgreSQL data across container rebuilds. - Done
   d. Configure environment variables for easy connection between backend and DB. - Done

3. Configure `.devcontainer/devcontainer.json` - Done
   a. Add essential VSCode extensions for both environments (e.g., `ms-python.python`, `ms-python.vscode-pylance`, `dbaeumer.vscode-eslint`, `esbenp.prettier-vscode`). - Done
   b. Set up port forwarding (`8000` for backend API, `5173` for frontend Vite server, `5432` for DB). - Done
   c. Add a `postCreateCommand` to automate dependency installation. This script should use `uv venv` and `uv pip install` for the backend, and run `pnpm install` in the frontend directory. - Done

4. Adapt Backend Database Management - Done
   a. Modify `src/backend/src/database/docker_utils.py` and `src/backend/src/database/chronocone_db_manager.py` to recognise when it's running inside a DevContainer with a linked `db` service. - Done
   b. Ensure the database initializer connects directly to the `db` service instead of trying to spawn a new Docker container using the Docker SDK. - Done

5. Documentation Updates - Done
   a. Update `DEVELOPMENT.md` to emphasize the DevContainer workflow as the primary development approach. - Done
   b. Update `README.md` Quick Start guide to include instructions on opening via VSCode DevContainers. - Done

## Notes

- This story shifts the development paradigm from requiring multiple local tools (Python, Node, Docker) to requiring *only* Docker and an IDE.
- Centralizing database orchestration in `docker-compose.yml` within the `.devcontainer` is generally preferred over a custom Python script attempting to spin up sidecar containers during development.
- Bootstrapping of project dependencies is handled by the `postCreateCommand` lifecycle hook in `devcontainer.json` because attempting to copy nested `requirements.txt` or `package.json` files during image build is brittle and requires precise context paths.
