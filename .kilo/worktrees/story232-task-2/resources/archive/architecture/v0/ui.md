Directives for agentic or AI writers and editors:

1. You MUST leave this list of directives in place as the first item in the document and unaltered as changes are made to this document.
2. When editing this document you MUST use either the Technical Writer or Technical Editor roles. If you are not in one of these modes you MUST NOT edit this document and inform the user of the requirement.
3. This document contains annotations that are labelled as "Notes:" or "Questions:" and you MUST NOT edit these annotations. You should move them to keep them with the content they relate to.

--- Begin Document Content ---

# Web UI

## Purpose

The Web UI is the primary mechanism for providing a visual display of metrics through a browser-based interface. It serves as the main interface for end users who need to interact with the ChronoCone system, providing access to both the data produced by the planning process and the metadata that defines how the system operates.

The Web UI bridges the gap between the technical complexity of the backend services and the practical needs of business users. While the Metrics Service and compute engine handle sophisticated data processing, the Web UI presents this information in an accessible, intuitive form. Users can explore metrics, understand relationships, and manage metadata without needing to understand the underlying technical architecture.

## Target Users

End users interacting with the ChronoCone system represent the primary audience for the Web UI. These users include business analysts who need to explore planning data, planners who need to review and adjust metadata, and managers who need to understand the current state of the planning process. What these users share is a need to access and understand planning data without requiring technical expertise in databases or programming.

The Web UI is designed to accommodate different levels of user sophistication. Casual users can navigate through pre-built views and explore data through intuitive interactions. Power users can access more advanced features for metadata management and detailed data exploration. This tiered approach ensures that the Web UI remains accessible while still providing the capabilities that advanced users need.

## Responsibilities

The Web UI has two primary responsibilities that together provide comprehensive support for user interaction with the ChronoCone system.

**Displaying metrics** involves presenting time series data in forms that users can understand and act upon. This includes showing forecasts and actuals, comparing values across different dimensions and time periods, and highlighting trends and anomalies. The display must handle the volume and complexity of planning data while remaining responsive and intuitive.

**Displaying and updating metadata** provides users with access to the structural information that defines how the system operates. Users can view categories, elements, metrics, and other metadata types, and authorized users can create and modify these definitions. This capability is essential for maintaining the planning system as business requirements evolve.

## Key Features

### Navigation Structure

The Web UI implements a top-level navigation bar that remains visible across all pages. This persistent navigation provides a consistent frame of reference for users as they move through different parts of the application, reducing cognitive load and making it easy to switch between contexts.

The navigation bar provides access to three main sections that together cover the primary use cases for the Web UI.

**Home** serves as the landing page with a welcome message. This page orients users to the application and provides a starting point for navigation. Future versions may include dashboard-style summaries of key metrics or recent activity.

**Definitions** provides the metadata management interface. This section allows users to view and manage the structural definitions that underpin the planning system, including categories, elements, metrics, and other metadata types.

**Metrics** provides the metrics display page. This section allows users to explore the time series data produced by the planning process, viewing forecasts and actuals across different dimensions and time periods.

### Page Structure

#### Home Page

The Home page serves as the application landing page and displays a welcome message to users. While currently simple, this page is designed to be extended with dashboard functionality in future versions. The welcome message provides context for new users and confirms that they have successfully accessed the system.

#### Definitions Page

The Definitions page provides an interface for viewing and managing metadata types. This page is central to the administrative functions of the ChronoCone system, allowing authorized users to define and modify the structures that govern how planning works.

The page includes several key features that support the metadata management workflow. **Metadata Type Tables** display metadata organized by type (categories, namespaces, units, etc.) in tabular format. This organization makes it easy to find and compare items within a type. **Add Functionality** provides a "+" icon button for each metadata type section to add new definitions. This inline addition capability streamlines the workflow for creating new metadata items. **Inline Editing** allows new rows to be added to tables for creating metadata entries directly within the table view, reducing context switching and making the creation process more intuitive.

#### Metrics Page

The Metrics page provides the primary interface for exploring time series data. It displays metrics with their actuals and forecasts, organized according to the calculation dependency structure.

##### Page Header Controls

At the top of the Metrics page, three dropdown controls provide context selection:

**Namespace Family Dropdown**: Selects the namespace family to view. This determines the high-level organizational context for the metrics display.

**Namespace Dropdown**: Selects the specific namespace within the chosen family. This determines the specific planning context (e.g., development, production, personal).

**Published/Unpublished Toggle**: Controls whether to display only published metrics or to include unpublished metrics as well. Options include:

- "Published Only" (default): Shows only metrics that have been approved for production use.
- "Published + Unpublished": Shows both published metrics and any unpublished versions that exist.

##### Time Series Display

Each metric appears as a horizontal time series with the following structure:

**Layout**: The time series flows from left to right chronologically. A vertical dividing line separates the actuals region (left) from the forecasts region (right). This division reflects the fundamental distinction between historical measurements and future predictions.

**Time Labels**: Above each cell in the time series, a time label appears in small, greyed out, italicized characters. The format of these labels depends on configuration:

| Time Grain | Default Format                | Alternative Format |
| ---------- | ----------------------------- | ------------------ |
| Weekly     | Week number (e.g., "Week 12") | YYYY-MM-DD         |
| Daily      | Day number (e.g., "Day 85")   | YYYY-MM-DD         |

The Metrics Service provides an administrative endpoint for configuring the default time label format per metric.

**Cell Values**: Each cell displays the metric value for that time period. The visual styling may vary based on whether the value is an actual or a forecast.

##### Vertical Arrangement

Metrics are arranged vertically on the page in directed acyclic graph (DAG) order:

- **Input metrics** (those that provide data to other metrics) appear at the top
- **Output metrics** (those that receive data from other metrics) appear at the bottom
- This arrangement reflects the flow of data through the calculation pipeline

The vertical ordering helps users understand dependencies and trace how values propagate through the system.

##### Unpublished Metric Display

When the user selects "Published + Unpublished" mode, unpublished metrics appear directly below their corresponding published metrics. This placement enables direct comparison between the current production state and pending changes.

Unpublished metrics are visually distinguished through:

1. **Color differentiation**: A distinct color scheme indicates unpublished status
2. **Pattern or icon overlay**: An additional visual indicator ensures accessibility for color-blind users

This dual-indicator approach ensures that all users can distinguish between published and unpublished metrics regardless of visual abilities.

##### Metric Detail View

Clicking on a metric opens a detail view with additional information:

**Graph Panel (All Users)**: Displays a visualization showing:

- Historical actuals versus historical forecasts (demonstrating forecast accuracy)
- Future forecasts (showing predicted outcomes)

This graph helps users understand both how well predictions have matched reality and what the system expects to happen going forward.

**Transformer Definition Panels (Plan Modification Users)**: Users with plan modification permission see additional panels displaying:

- **Forecast Transformer Definition**: The code and configuration that produces forecast values
- **Actuals Transformer Definition** (if applicable): The code and configuration that produces actual values from external data sources

These panels help planners understand how metric values are calculated and identify opportunities for improvement.

##### Component Architecture

The Metrics page consists of several React components:

**MetricsPage** (`src/pages/Metrics.tsx`): The main page component that orchestrates the metrics display. It manages state for namespace selection, published/unpublished toggle, and selected metric for detail view.

**NamespaceSelector**: A composite component containing the namespace family and namespace dropdowns. It handles loading of available namespaces and user selection.

**PublishedToggle**: A dropdown component for selecting between "Published Only" and "Published + Unpublished" display modes.

**MetricsList**: A container component that displays metrics in vertical DAG order. It receives metric data from the parent and renders individual MetricRow components.

**MetricRow**: A component that displays a single metric as a horizontal time series. It handles:

- Rendering the time series cells with actuals and forecasts
- Displaying time labels above each cell
- Applying visual styling for published vs. unpublished status
- Handling click events to open the detail view

**MetricDetail**: A modal or panel component that displays when a user clicks on a metric. It shows the graph and, for authorized users, the transformer definition panels.

**TimeSeriesGraph**: A component that renders the actuals vs. forecast visualization using a charting library.

**TransformerPanel**: A component that displays transformer code with syntax highlighting and metadata annotations.

## Component Structure

The Web UI is built using React components that promote reusability and maintainability. Each component has a clear responsibility and can be developed and tested independently.

### NavigationBar Component

Located at `src/components/NavigationBar.tsx`, the NavigationBar provides top-level navigation between pages using React Router. This component is responsible for rendering the navigation bar that appears at the top of every page, handling navigation state, and highlighting the current section.

### Page Components

The page components correspond to the main sections of the application and are responsible for rendering the content of each page.

**Home** (`src/pages/Home.tsx`) is the welcome page component. It renders the landing page content and serves as the default destination when users first access the application.

**Definitions** (`src/pages/Definitions.tsx`) is the metadata management page with table components. It orchestrates the display of metadata across different types and provides the interface for creating and modifying metadata items.

**Metrics** (`src/pages/Metrics.tsx`) is the placeholder page for future metrics display. It currently renders a simple message but will be expanded to provide comprehensive time series visualization capabilities.

### MetadataTable Component

Located at `src/components/MetadataTable.tsx`, the MetadataTable displays metadata entries in a tabular format with support for adding new rows. This component is designed to be reusable across different metadata types, accepting configuration that specifies which columns to display and how to handle user interactions.

The component handles the complexity of tabular data display, including sorting, filtering, and inline editing. By encapsulating this functionality in a reusable component, the Web UI maintains consistency across different metadata types while reducing code duplication.

## Related Documentation

- [v0 Architecture Overview](v0_architecture.md)
- [v0 Requirements - Web UI Section](../requirements/v0_requirements/v0_requirements.md#35-web-ui)
- [Story007: UI Navigation and Pages](../stories/Story007_ui-navigation-and-pages.md) - Implementation story for initial UI structure

---

# Plugin UI

## Purpose

The Plugin UI provides the user interface for planners working on transformers, shifters, and filters through IDE plugins. Unlike the Web UI, which is browser-based, the Plugin UI is integrated directly into the developer's IDE, providing a seamless development experience for creating and modifying calculation logic.

The Plugin UI bridges the gap between the Metrics Service and the developer's local workspace. While the Metrics Service stores and manages calculation assets in the object store, the Plugin UI provides the interface for retrieving, editing, testing, and pushing those assets. This integration enables a workflow that combines AI-assisted development with traditional code editing.

## Target Users

Planners and developers who create or modify transformers, shifters, and filters. These users need to work with calculation logic in a development environment that provides code editing, testing, and AI assistance capabilities.

## Supported Platforms

The Plugin UI is implemented as two separate plugins:

- **IntelliJ Plugin**: For PyCharm IDE (and compatible JetBrains IDEs). This plugin is built using the IntelliJ Platform SDK and integrates with PyCharm's Python development features.

- **VS Code Extension**: For Visual Studio Code. This extension is built using the VS Code Extension API and integrates with VS Code's Python extension ecosystem.

Both plugins provide identical functionality and user experience, differing only in their implementation details to match their respective platforms.

## Responsibilities

1. Provide an integrated interface for the pull-modify-push workflow
2. Display and edit transformer, shifter, and filter code
3. Provide access to the agentic interface for AI-assisted development
4. Display test results and validation feedback
5. Manage local sandbox state and synchronization with the Metrics Service

## Architecture

### Docker-Based Sandbox

The Plugin UI implements the sandboxed execution environment using Docker containers. When a user connects to a namespace family and namespace, the plugin creates or connects to a Docker container that serves as the isolated execution environment for that namespace.

The Docker container provides several critical capabilities:

**Isolation**: Code execution occurs within the container, completely isolated from the user's local machine. This isolation ensures that the code and its dependencies do not interfere with the user's local environment, and vice versa.

**Consistent Runtime**: The container provides a standardized Python runtime with known versions of Python and required libraries. This consistency ensures that functions behave the same way during testing as they will in production calculations.

**Namespace Connection**: The container is configured with the code, packages, and configuration appropriate for the connected namespace family and namespace. Some of this content comes from the object store, including the transformers, shifters, filters, and their associated test data.

**Dependency Management**: The container includes all packages and programs required to execute the calculation logic. This eliminates the need for users to manage Python environments or install dependencies on their local machines.

### Connection Flow

When a user connects to a namespace family and namespace, the following sequence occurs:

1. The plugin checks whether a Docker container for that namespace already exists. If not, it creates one from a base image.

2. The plugin retrieves the code and configuration for the namespace from the Metrics Service. This includes transformers, shifters, filters, and their associated test data.

3. The plugin synchronizes the retrieved content into the Docker container, ensuring the container has the current state of all calculation assets.

4. The plugin updates its UI to reflect the connected state, displaying the explorer panel with the metrics, shifters, and filters from the namespace.

### Push/Pull Operations

The Plugin UI implements push and pull operations that synchronize code and test data between the local workspace (Docker container) and the object store.

**Pull Operation**: When a user pulls a transformer, shifter, or filter, the plugin retrieves both the Python code and its associated test data from the object store. The test data is stored in Parquet format in the object store, and the plugin converts it to the tabular format displayed in the test panel.

**Push Operation**: When a user pushes a modified transformer, shifter, or filter, the plugin sends both the updated Python code and the current test data to the object store. The test data is converted from the tabular format to Parquet format for storage.

This co-location of code and test data ensures that they remain synchronized. When code is versioned, its associated test data is versioned alongside it.

## Key Features

### Namespace Connection Panel

When a user opens the plugin, the first interaction is namespace selection. The plugin displays a connection panel that prompts the user to connect to a namespace family and namespace combination. This connection establishes the context for all subsequent operations.

If already connected, the panel displays the current connection status with an option to switch to a different namespace.

### Explorer Panel

The explorer panel provides navigation through calculation assets in the connected namespace. It contains three vertically arranged tabs with expand/collapse controls:

- **Metrics Tab**: Displays metrics arranged by their DAG dependencies, flowing from upstream (top) to downstream (bottom).

- **Shifters Tab**: Displays shifters with their dimensional transformation descriptions.

- **Filters Tab**: Displays filters with their split descriptions.

### Chat Window

A chat window at the bottom of the interface provides interaction with AI agents. Users can request code generation, modification, and explanation through natural language prompts. The AI agents can also manipulate and generate test data.

### Detail Views

When a user selects an asset from the explorer panel, a detail view opens to the right:

- **Metric Detail View**: Shows the transformers (forecast and optionally actuals) with greyed-out annotations indicating dimensions and value types.

- **Shifter/Filter Detail View**: Shows the Python code implementing the function.

### Testing Interface

Each function displays a test icon that opens a test panel with input/output test data. Users can run tests, add and delete test rows, and edit test values. Test results show passed (green) or failed (red) status with actual output displayed for failures.

## Component Structure

### IntelliJ Plugin Components

The IntelliJ plugin is structured using the IntelliJ Platform SDK's standard component model:

- **Tool Window**: The main plugin panel that appears in the IDE's tool window area.
- **Actions**: Menu items and toolbar buttons for namespace connection and other operations.
- **Services**: Background services for Docker container management and Metrics Service communication.
- **Editors**: Custom editors for transformer, shifter, and filter code with annotation support.

### VS Code Extension Components

The VS Code extension follows the VS Code extension architecture:

- **Webview Panel**: The main plugin panel rendered as a webview.
- **Commands**: VS Code commands for namespace connection and other operations.
- **Language Client**: Integration with the Python language server for code intelligence.
- **Terminal Integration**: For Docker container management and test execution.

## Related Documentation

- [v0 Architecture Overview](v0_architecture.md)
- [v0 Requirements - Plugin UI Section](../requirements/v0_requirements/v0_requirements.md#38-plugin-ui)
- [v0 Requirements - IDE Plugins Section](../requirements/v0_requirements/v0_requirements.md#37-ide-plugins)
- [Data Storage](data_storage.md) - For test data Parquet storage format
