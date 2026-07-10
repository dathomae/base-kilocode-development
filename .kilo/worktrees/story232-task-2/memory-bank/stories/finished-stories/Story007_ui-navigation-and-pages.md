# Story007: UI Navigation and Pages

## Goal

Replace the default React application with a new application that includes a navigation bar and three main pages: Home, Definitions, and Metrics. The Home page displays a welcome message. The Definitions page shows a table of metadata types with the ability to add new definitions. The Metrics page displays a placeholder message for future implementation.

## References

- [Architecture](../specifications/architecture/)
- [Requirements](../specifications/requirements/)
- [Design](../specifications/design/)

## Dependencies

- [Story003_frontend-structure.md](Story003_frontend-structure.md) - Done (must complete React project setup first)

## Dependent Stories

(To be populated as other stories are added)

## Tasks

1. Create navigation bar component - Completed
   a. Create `src/components/NavigationBar.tsx` component with Home, Definitions, and Metrics links - Completed
   b. Style the navigation bar to appear at the top of the page - Completed
   c. Add navigation state management using React Router - Completed
   d. Create `src/components/NavigationBar.test.tsx` with tests for rendering and navigation - Completed
   e. Run tests to verify navigation bar renders correctly - Completed

2. Set up React Router for page navigation - Completed
   a. Install React Router dependency (`npm install react-router-dom`) - Completed
   b. Create route configuration in `src/App.tsx` for Home, Definitions, and Metrics pages - Completed
   c. Update `src/App.test.tsx` to verify routing works correctly - Completed
   d. Run tests to verify route navigation - Completed

3. Create Home page component - Completed
   a. Create `src/pages/Home.tsx` component with welcome message - Completed
   b. Style the Home page with appropriate layout and typography - Completed
   c. Create `src/pages/Home.test.tsx` with tests for welcome message display - Completed
   d. Run tests to verify Home page renders correctly - Completed

4. Create Definitions page with metadata type tables - Completed
   a. Create `src/pages/Definitions.tsx` component - Completed
   b. Define metadata types (categories, namespaces, units, etc.) based on schema - Completed
   c. Create table component `src/components/MetadataTable.tsx` for displaying metadata - Completed
   d. Implement table with columns for metadata type, name, and description - Completed
   e. Group definitions by metadata type in separate table sections - Completed
   f. Add "+" icon button for each metadata type section to add new rows - Completed
   g. Implement add row functionality that creates editable row in the table - Completed
   h. Create `src/pages/Definitions.test.tsx` with tests for table display and add functionality - Completed
   i. Run tests to verify Definitions page works correctly - Completed

5. Create Metrics placeholder page - Completed
   a. Create `src/pages/Metrics.tsx` component with "some day soon" message - Completed
   b. Style the Metrics page appropriately - Completed
   c. Create `src/pages/Metrics.test.tsx` with tests for message display - Completed
   d. Run tests to verify Metrics page renders correctly - Completed

6. Update main App component - Completed
   a. Replace default Vite/React content in `src/App.tsx` with new application structure - Completed
   b. Integrate NavigationBar component at the top level - Completed
   c. Set up page routing with React Router - Completed
   d. Update `src/App.test.tsx` to test the new application structure - Completed
   e. Run all tests to verify complete application works - Completed

7. Clean up default application files - Completed
   a. Remove unused default assets (React and Vite logos) - Completed
   b. Remove or update default CSS styles that are no longer needed - Completed
   c. Verify no broken imports or references remain - Completed
   d. Run tests to ensure cleanup did not break functionality - Completed

8. Verify complete UI implementation - Completed
   a. Run the development server (`npm run dev`) - Completed
   b. Manually verify navigation between all three pages works - Completed
   c. Verify Home page displays welcome message - Completed
   d. Verify Definitions page shows tables and add functionality works - Completed
   e. Verify Metrics page displays placeholder message - Completed
   f. Run full test suite to ensure all tests pass - Completed
   g. Build production bundle to verify no build errors - Completed

## Notes

- The Definitions page should display metadata types based on the taxonomy schema defined in `src/backend/schema/taxonomy.sql`
- The add row functionality in Definitions should create an editable row but does not need to persist to backend in this story (that will come in a later story)
- Styling should be clean and functional; detailed UI polish will be addressed in future stories
- The navigation bar should remain visible when navigating between pages
