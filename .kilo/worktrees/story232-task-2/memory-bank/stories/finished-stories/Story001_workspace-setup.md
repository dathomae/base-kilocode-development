# Story001: Basic Multi-Root Workspace Setup

## Goal

Set up CHRONOCONE as a multi-root workspace with properly configured backend (Python) and frontend (React/Node) environments. This includes configuring the workspace structure, creating isolated development environments for each component, and adjusting the CI/CD pipeline to work with the multi-root architecture.

## References

- [Brief](../brief.md)
- [Requirements](../specifications/requirements/)

## Dependencies

None - This is the foundational story for the project.

## Dependent Stories

- [Story002_backend-structure.md](Story002_backend-structure.md) - Done (requires backend environment setup)
- [Story003_frontend-structure.md](Story003_frontend-structure.md) - Done (requires frontend environment setup)

## Tasks

1. Create VSCode multi-root workspace configuration - Completed
   a. Create `.vscode/chronocone.code-workspace` file defining three workspace roots: the top-level CHRONOCONE, `src/backend`, and `src/frontend` - Completed
   b. Configure workspace-specific settings for each root (e.g., Python interpreter for backend, Node/npm settings for frontend) - Completed
   c. Test opening the workspace to verify all roots are recognized - Completed

2. Set up Python backend environment - Completed
   a. Navigate to `src/backend` and create a Python virtual environment using `python -m venv .venv` - Completed
   b. Create basic `requirements.txt` in the backend directory with minimal initial dependencies (to be expanded in Story002) - Completed
   c. Activate the virtual environment and install dependencies - Completed
   d. Verify Python environment is properly configured and can execute basic Python commands - Completed
   e. Create a basic `main.py` file in the backend directory with a simple FastAPI "Hello World" endpoint for CI verification - Completed

3. Set up React/Node frontend environment - Not Started
   a. Navigate to `src/frontend` and initialize a new React project using appropriate tooling (e.g., Vite or Create React App) - Not Started
   b. Verify that `package.json` and `node_modules` are properly created - Not Started
   c. Run the default development server to ensure the frontend environment is working - Not Started
   d. Create a simple test to verify the React environment is working correctly - Not Started

4. Update CI/CD configuration for multi-root workspace - Completed
   a. Examine the existing `.github/workflows/ci.yml` file to understand current configuration - Completed
   b. Modify the CI workflow to install Python dependencies from `src/backend/requirements.txt` - Completed
   c. Add steps to install npm dependencies from `src/frontend/package.json` - Completed
   d. Configure basic environment setup for backend and frontend (detailed test configuration will be done in later stories) - Completed
   e. Test the CI pipeline by pushing changes and verifying basic environment setup works - Completed

5. Update top-level project configuration - Completed
   a. Review and update top-level `.gitignore` to properly exclude `.venv`, `node_modules`, and other environment-specific artifacts from both backend and frontend - Completed
   b. Update or create top-level `README.md` to document the multi-root workspace structure and setup instructions - Completed
   c. Document environment setup procedures for new developers - Completed

## Notes

- The backend will use Python 3.11+ with FastAPI for the web framework
- The frontend will use React 18+ with TypeScript for type safety
- Both environments should have their own isolated dependency management
- This story focuses on basic environment setup; detailed project structure (src/tests directories) and testing infrastructure will be handled in Story002
