# CHRONOCONE Frontend

A React + TypeScript frontend application for CHRONOCONE, built with Vite.

## Technology Stack

- **Framework**: React 19
- **Language**: TypeScript (strict mode enabled)
- **Build Tool**: Vite 6
- **Testing**: Vitest with React Testing Library
- **Linting**: ESLint

## Project Structure

```
src/
├── components/       # Reusable React components
├── config.ts         # Environment variable configuration
├── assets/           # Static assets (images, fonts)
├── App.tsx           # Main application component
├── main.tsx          # Application entry point
├── setupTests.ts     # Test configuration
└── ...
```

## Getting Started

### Prerequisites

- Node.js (version 18 or higher recommended)
- npm

### Installation

1. Install dependencies:

```bash
npm install
```

### Environment Variables

The application uses environment variables for configuration. These variables must be prefixed with `VITE_` to be exposed to the client-side code (as required by Vite).

#### Setup

1. Copy the example environment file:

```bash
cp .env.example .env
```

2. Edit `.env` and update the values as needed:

```env
# Backend Core Server URL
VITE_CORE_SERVER_URL=http://localhost:8000
```

#### Available Variables

| Variable                   | Description                     | Default                 |
| -------------------------- | ------------------------------- | ----------------------- |
| `VITE_CORE_SERVER_URL`     | URL for the backend API server  | `http://localhost:8000` |
| `VITE_METRICS_SERVICE_URL` | URL for the metrics service API | `http://localhost:8001` |

#### Using Environment Variables in Code

Import from the centralized config module:

```typescript
import { CORE_SERVER_URL, getApiUrl } from '@/config';

// Use the URL directly
console.log(CORE_SERVER_URL); // "http://localhost:8000"

// Or use the helper to construct API URLs
const statusUrl = getApiUrl('/api/v1/status');
// Result: "http://localhost:8000/api/v1/status"
```

**Note**: Never commit the `.env` file to version control. It is already listed in `.gitignore`.

### Using Environment Variables in Code

Import from the centralized config module:

```typescript
import { CORE_SERVER_URL, METRICS_SERVICE_URL, getApiUrl, getMetricsApiUrl } from '@/config';

// Use the URLs directly
console.log(CORE_SERVER_URL); // "http://localhost:8000"
console.log(METRICS_SERVICE_URL); // "http://localhost:8001"

// Or use the helpers to construct API URLs
const statusUrl = getApiUrl('/api/v1/status');
// Result: "http://localhost:8000/api/v1/status"

const typesUrl = getMetricsApiUrl('/types');
// Result: "http://localhost:8001/types"
```

## Services Integration

### Metrics Service

The frontend integrates with the ChronoCone Metrics Service for metadata management. The Definitions page fetches metadata types and items from the Metrics Service and persists new entries to the backend database.

#### API Service Module

The `src/services/metricsApi.ts` module provides typed functions for interacting with the Metrics Service:

- `fetchMetadataTypes()` - Retrieves supported metadata types from `/types` endpoint
- `fetchMetadataList(type?)` - Retrieves metadata items from `/list` endpoint
- `createMetadata(type, name, description)` - Creates new metadata via `/metadata` endpoint
- Type mapping functions for converting between frontend and backend type names

#### Configuration

Set the Metrics Service URL via environment variable:

```bash
VITE_METRICS_SERVICE_URL=http://localhost:8001
```

#### Troubleshooting

| Issue                             | Solution                                                     |
| --------------------------------- | ------------------------------------------------------------ |
| "Unable to connect to the server" | Ensure Metrics Service is running on port 8001               |
| "Failed to fetch metadata types"  | Check Metrics Service health endpoint at `/health/db`        |
| Empty metadata tables             | Verify database is initialized and contains metadata entries |

## Development

### Start Development Server

```bash
npm run dev
```

The application will be available at `http://localhost:5173` (or another port if 5173 is in use).

### Type Checking

```bash
npm run typecheck
```

### Linting

```bash
npm run lint
```

## Testing

### Run Tests

```bash
npm test
```

### Run Tests in CI Mode (non-interactive)

```bash
npm test -- --run
```

### Run Tests in Watch Mode

```bash
npm test -- --watch
```

## Building for Production

```bash
npm run build
```

The production build will be output to the `dist/` directory.

## Code Quality

### Formatting

This project uses Prettier for code formatting. Configuration is in [`.prettierrc`](.prettierrc):

| Option           | Value   | Description                                    |
| ---------------- | ------- | ---------------------------------------------- |
| `semi`           | `true`  | Add semicolons at end of statements            |
| `singleQuote`    | `true`  | Use single quotes instead of double            |
| `tabWidth`       | `2`     | Number of spaces per indentation               |
| `trailingComma`  | `es5`   | Trailing commas where valid in ES5             |
| `printWidth`     | `100`   | Wrap lines longer than 100 characters          |
| `bracketSpacing` | `true`  | Spaces between brackets in object literals     |
| `arrowParens`    | `avoid` | Omit parens when arrow function has single arg |

#### Format Commands

```bash
# Format all files
npm run format

# Check formatting without modifying files
npm run format:check
```

### Linting

ESLint is configured with TypeScript, React Hooks, and Prettier integration. See [`eslint.config.js`](eslint.config.js) for details.

```bash
# Run linter
npm run lint

# Run linter with auto-fix
npm run lint:fix
```

## Coding Conventions

### TypeScript

- **Strict mode enabled**: All TypeScript strict checks are active
- Use explicit types for function parameters and return values
- Prefer `interface` over `type` for object definitions
- Use path aliases (`@/`) for imports from `src/`

### React

- Use functional components with hooks
- Component files use PascalCase (e.g., `MyComponent.tsx`)
- Hook files use camelCase with `use` prefix (e.g., `useAuth.ts`)
- Place reusable components in `src/components/`

### Testing

- Co-locate tests with source files using `.test.tsx` suffix
- Use React Testing Library for component tests
- Follow AAA pattern: Arrange, Act, Assert

### File Organization

```
src/
├── components/       # Reusable React components
│   └── ComponentName/
│       ├── index.tsx
│       └── ComponentName.test.tsx
├── hooks/            # Custom React hooks
├── utils/            # Utility functions
├── types/            # Shared TypeScript types
├── config.ts         # Environment configuration
└── ...
```

## Path Aliases

The project is configured with path aliases for cleaner imports:

- `@/` maps to `./src/`

Example:

```typescript
// Instead of relative paths like:
import { Component } from '../../components/Component';

// Use:
import { Component } from '@/components/Component';
```

## Build and Deployment

### Production Build

Create an optimized production build:

```bash
npm run build
```

This command:

1. Runs TypeScript compiler to check for type errors
2. Bundles the application with Vite
3. Outputs to the `dist/` directory

### Preview Production Build

Test the production build locally:

```bash
npm run preview
```

This serves the `dist/` directory at `http://localhost:4173`.

### Build Output

The `dist/` directory contains:

- `index.html` - Entry point
- `assets/` - Bundled JS, CSS, and other assets (with hashed filenames for caching)

### Environment-Specific Builds

To build for different environments:

```bash
# Development (includes source maps)
npm run build -- --mode development

# Production (optimized, minified)
npm run build
```

### CI/CD Integration

The project includes GitHub Actions workflow (`.github/workflows/ci.yml`) that:

1. Installs dependencies
2. Runs TypeScript type checking
3. Runs ESLint
4. Executes Vitest tests
5. Builds the production bundle

## Learn More

- [Vite Documentation](https://vitejs.dev/)
- [React Documentation](https://react.dev/)
- [Vitest Documentation](https://vitest.dev/)
