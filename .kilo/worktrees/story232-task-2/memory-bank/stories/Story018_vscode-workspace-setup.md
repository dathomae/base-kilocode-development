# Story 018: VSCode Extension Workspace Setup

## Overview

Set up the VSCode extension workspace under the chronochrome directory. This includes creating the basic project structure for a TypeScript-based VSCode extension, configuring the build system with npm/webpack, and establishing the development environment for the ChronoCone VSCode extension.

## References

- [Architecture](../specifications/architecture/)
- [VSCode Extension API Documentation](https://code.visualstudio.com/api)
- [Story001_workspace-setup.md](Story001_workspace-setup.md) - Original workspace setup for reference

## Dependencies

### Prerequisites

- [Story001_workspace-setup.md](Story001_workspace-setup.md) - Done (base workspace structure)

### Dependents

- [Story019_timeseries-parser-typescript.md](Story019_timeseries-parser-typescript.md) - Not Started (requires VSCode extension project structure)

## Tasks

1. Create VSCode extension project directory structure - Not Started
   a. Create `chronochrome/vscode/` directory under project root
   b. Create `src/` directory for TypeScript source files
   c. Create `src/test/` directory for extension tests
   d. Create `media/` directory for webview resources
   e. Create `resources/` directory for extension resources

2. Initialize npm project and configure dependencies - Not Started
   a. Create `package.json` with VSCode extension configuration
   b. Configure extension activation events and contributions
   c. Add TypeScript and webpack as development dependencies
   d. Add VSCode extension test framework (@vscode/test-electron)
   e. Add other required dependencies (Parquet library, etc.)

3. Configure TypeScript build - Not Started
   a. Create `tsconfig.json` with strict TypeScript configuration
   b. Configure compilation target for VSCode extension host
   c. Set up source maps for debugging
   d. Configure module resolution
   e. Create type definitions for any untyped dependencies

4. Configure webpack bundling - Not Started
   a. Create `webpack.config.js` for extension bundling
   b. Configure entry point and output directory
   c. Set up development and production build modes
   d. Configure source map generation
   e. Add webpack build scripts to package.json

5. Create extension configuration files - Not Started
   a. Create `.vscode/launch.json` for extension debugging
   b. Create `.vscode/tasks.json` for build tasks
   c. Create `.vscodeignore` for packaging exclusions
   d. Create `vsc-extension-quickstart.md` for development guide
   e. Configure extension icon and display settings

6. Set up basic extension structure - Not Started
   a. Create `src/extension.ts` as extension entry point
   b. Implement `activate()` and `deactivate()` functions
   c. Create basic command registration
   d. Create `src/test/suite/extension.test.ts` for unit tests
   e. Verify extension activates in VSCode development host

7. Configure development environment - Not Started
   a. Add `npm run compile` script for TypeScript compilation
   b. Add `npm run watch` script for development
   c. Add `npm run test` script for running tests
   d. Add `npm run package` script for extension packaging
   e. Document development setup in README.md

8. Add project to multi-root workspace - Not Started
   a. Update root workspace configuration to include vscode folder
   b. Configure workspace-specific settings for TypeScript
   c. Verify workspace opens correctly in VSCode
   d. Document workspace setup in project README

9. Verify build and test infrastructure - Not Started
   a. Run `npm run compile` to verify TypeScript compiles
   b. Run `npm run test` to verify test framework works
   c. Run `npm run package` to verify extension packages correctly
   d. Press F5 to verify extension loads in development host
   e. Document build commands in README.md

## Decisions

1. **TypeScript configuration**: Use strict TypeScript configuration with `strict: true` to ensure type safety throughout the extension codebase.

2. **Webpack bundling**: Use webpack to bundle the extension. This reduces extension size and improves load time compared to unbundled TypeScript output.

3. **Extension activation**: Use `onCommand` and `onLanguage` activation events to ensure the extension only loads when needed.

4. **Test framework**: Use the VSCode extension test framework (`@vscode/test-electron`) for integration tests and Mocha for unit tests.

5. **Package structure**: Use a flat `src/` directory structure with subdirectories for specific features (e.g., `src/timeseries/`, `src/views/`).

## Notes

- The VSCode Extension API documentation should be consulted for best practices on extension structure and configuration.
- Consider using the VSCode Extension Generator (`yo code`) as a starting point for project structure.
- The extension should be designed to work with both VSCode and VSCode Insiders editions.
- Future stories will add time series parsing functionality and editor support.
- Consider setting up continuous integration for the extension build.
- The extension will need to handle large Parquet files efficiently, so consider Web Workers for background processing.
