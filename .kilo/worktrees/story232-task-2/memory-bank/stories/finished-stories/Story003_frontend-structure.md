# Story003: Frontend React/TypeScript Project Setup

## Goal

Set up the frontend workspace as a React project with TypeScript using Vite as the build tool and Vitest as the testing framework. This includes generating the default React application structure, configuring TypeScript, setting up Vitest testing infrastructure, and verifying that the development and testing environments work correctly.

## References

- [Architecture](../specifications/architecture/)
- [Brief](../brief.md)

## Dependencies

- [Story001_workspace-setup.md](Story001_workspace-setup.md) - Completed (must complete basic environment setup first)

## Dependent Stories

(To be populated as other stories are added)

## Tasks

1. Initialize React project with TypeScript using Vite - Completed
   a. Navigate to `src/frontend` directory - Completed
   b. Initialize a new React project using Vite with TypeScript template (`npm create vite@latest . -- --template react-ts`) - Completed
   c. Verify that `package.json`, `tsconfig.json`, `vite.config.ts`, and project structure are properly created - Completed
   d. Review generated files and folder structure to understand the default Vite + React setup - Completed

2. Install dependencies and verify basic setup - Completed
   a. Run `npm install` to install all dependencies - Completed
   b. Verify that `node_modules` directory is created and populated - Completed
   c. Review `package.json` scripts to understand available commands (dev, build, test, etc.) - Completed
   d. Start the development server (`npm run dev`) to verify the default app runs - Completed
   e. Access the default app in browser and verify it displays correctly - Completed
   f. Stop the development server - Completed

3. Configure and verify TypeScript setup - Completed
   a. Review `tsconfig.json` to understand TypeScript configuration - Completed
   b. Enable TypeScript strict mode for better type safety - Completed
   c. Add path aliases to `tsconfig.json` (e.g., `@/*` mapping to `./src/*`) - Completed
   d. Update `vite.config.ts` to resolve the path aliases for the build process - Completed
   e. Create a simple TypeScript component to verify type checking and path aliases work - Completed
   f. Run TypeScript compiler check (`tsc --noEmit`) to verify no type errors - Completed

4. Set up Vitest testing infrastructure - Completed
   a. Install Vitest and testing dependencies (`vitest`, `@testing-library/react`, `@testing-library/jest-dom`, `jsdom`) - Completed
   b. Update `vite.config.ts` to include Vitest configuration with jsdom environment - Completed
   c. Create test setup file `src/setupTests.ts` to configure testing library - Completed
   d. Update `package.json` to include test script (`"test": "vitest"`) - Completed
   e. Verify Vitest configuration by running `npm test -- --version` - Completed

5. Create initial test and verify Vitest execution - Completed
   a. Create `src/App.test.tsx` with a simple Vitest test for the default App component using React Testing Library - Completed
   b. Run tests (`npm test`) to verify Vitest test discovery and execution works - Completed
   c. Verify that Vitest output shows the test was discovered and passed - Completed
   d. Create a second test that intentionally fails to verify Vitest properly reports failures - Completed
   e. Fix the failing test to ensure baseline is all tests passing - Completed
   f. Run tests in watch mode (`npm test -- --watch`) to verify watch mode works correctly - Completed

6. Update top-level CI configuration - Completed
   - [x] a. Update `.github/workflows/ci.yml` to run npm install from the correct directory (`src/frontend`) - Completed
   - [x] b. Add CI step to run TypeScript type checking - Completed
   - [x] c. Add CI step to run Vitest tests in CI mode (`npm test -- --run`) - Completed
   - [x] d. Add CI step to verify frontend builds successfully (`npm run build`) - Completed
   - [x] e. Push changes to trigger CI workflow - Completed
   - [x] f. Verify that CI workflow runs successfully and all frontend checks pass - Completed
   - [x] g. If CI fails, diagnose and fix issues, then verify CI passes - Completed

7. Configure environment variables for backend connection - Completed
   a. Create `.env.example` file with template for backend core server URL (e.g., `VITE_CORE_SERVER_URL=http://localhost:8000`) - Completed
   b. Create `.env` file for local development with actual backend URL - Completed
   c. Verify `.env` is in `.gitignore` to prevent committing secrets - Completed
   d. Update Vite configuration if needed to ensure environment variables are properly loaded - Completed
   e. Create a simple config module (e.g., `src/config.ts`) to centralize environment variable access - Completed
   f. Document environment variable setup in README - Completed

8. Configure linting and code quality tools - Completed
   a. Verify ESLint is configured (usually included in Vite React template) - Completed
   b. Review and adjust ESLint configuration if needed - Completed
   c. Add Prettier for code formatting if not already included - Completed
   d. Run linter to verify no issues with generated code - Completed
   e. Add lint script to CI workflow - Completed

9. Document frontend setup and development procedures - Completed
   a. Create or update `src/frontend/README.md` with setup instructions - Completed
   b. Document how to run the development server locally - Completed
   c. Document how to run tests and linting - Completed
   d. Document project structure and coding conventions - Completed
   e. Document build and deployment procedures (if applicable) - Completed

10. Clean up and verify complete frontend setup - Completed
    a. Review all generated files and remove any unnecessary boilerplate - Completed
    b. Update default app content to be specific to CHRONOCONE (replace generic Vite/React branding) - Completed
    c. Run full test suite to ensure everything passes - Completed
    d. Build the production bundle to verify build process works - Completed
    e. Verify CI workflow passes for all frontend tasks - Completed

## Notes

- Using Vite for modern React projects provides faster development experience with instant HMR (Hot Module Replacement)
- TypeScript strict mode is enabled for better type safety
- Path aliases (e.g., `@/*` for `./src/*`) are configured in tsconfig.json for cleaner imports
- Vitest is used for testing due to its seamless integration with Vite and shared configuration
- Vitest provides fast test execution and watch mode for efficient development
- Environment variables are configured for backend connection (VITE_CORE_SERVER_URL)
- Vite requires environment variables to be prefixed with `VITE_` to be exposed to client code
- The frontend will eventually need to connect to the backend API, but actual API integration will be handled in later stories
