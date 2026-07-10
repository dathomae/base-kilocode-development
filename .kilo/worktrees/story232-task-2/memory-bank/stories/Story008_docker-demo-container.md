# Story008: Docker Demo Container

## Goal

Create a Docker container that packages the complete ChronoCone application for demonstrations. The container includes the PostgreSQL database server, automatically starts the Core Service on container startup, serves the frontend through a web server, and initializes the database with definitions from a JSON configuration file. When started, the container is immediately ready for use without any additional configuration.

## References

- [Architecture](../specifications/architecture/)
- [Design](../specifications/design/)

## Dependencies

- [Story004_core-and-metrics-services.md](Story004_core-and-metrics-services.md) - Done (requires Core Service and Metrics Service implementations)
- [Story005_database-initialization-cli.md](Story005_database-initialization-cli.md) - Done (requires database initialization capabilities)
- [Story007_ui-navigation-and-pages.md](Story007_ui-navigation-and-pages.md) - Done (requires frontend build artifacts)

## Dependent Stories

(To be populated as other stories are added)

## Tasks

1. Create Dockerfile for the demo container - Not Started
   a. Use multi-stage build with Go base image for the Core service - Not Started
   b. Install PostgreSQL server within the container - Not Started
   c. Install Node.js and build frontend in build stage - Not Started
   d. Configure nginx as the web server for serving frontend static files - Not Started
   e. Set up reverse proxy in nginx to forward API calls to backend services - Not Started
   f. Build the Go binary and include necessary assets - Not Started
   g. Copy backend source code (Go) into the container - Not Started
   h. Expose necessary ports (80 for web, 8000 for Core Service, 5432 for PostgreSQL if needed externally) - Not Started

2. Create container startup script - Not Started
   a. Create `scripts/container-startup.sh` that orchestrates service startup - Not Started
   b. Initialize PostgreSQL data directory if not already present - Not Started
   c. Start PostgreSQL server in the background - Not Started
   d. Wait for PostgreSQL to be ready for connections - Not Started
   e. Run database schema initialization using existing CLI - Not Started
   f. Load definitions from JSON configuration file into the database - Not Started
   g. Start the Core Service with autostart enabled - Not Started
   h. Start nginx to serve the frontend - Not Started
   i. Implement graceful shutdown handling for all services - Not Started

3. Create JSON configuration file for demo definitions - Not Started
   a. Create `config/demo-definitions.json` with sample metadata definitions - Not Started
   b. Define categories for demonstration (e.g., Sales, Inventory, Operations) - Not Started
   c. Define namespaces for demonstration (e.g., Retail, Manufacturing) - Not Started
   d. Define units for demonstration (e.g., units, dollars, percentages) - Not Started
   e. Define sample policies for demonstration - Not Started
   f. Define sample connectors for demonstration - Not Started
   g. Create JSON schema to validate the definitions file structure - Not Started

4. Create database initialization module for JSON loading - Not Started
   a. Create `internal/database/loader.go` - Not Started
   b. Implement function to parse JSON configuration file - Not Started
   c. Implement function to insert categories into the database - Not Started
   d. Implement function to insert namespaces into the database - Not Started
   e. Implement function to insert units into the database - Not Started
   f. Implement function to insert policies into the database - Not Started
   g. Implement function to insert connectors into the database - Not Started
   h. Add idempotency checks to prevent duplicate entries on container restart - Not Started
   i. Add error handling with clear messages for malformed JSON - Not Started

5. Configure nginx as the web server - Not Started
   a. Create `docker/nginx.conf` with server configuration - Not Started
   b. Configure root location to serve frontend static files from `/usr/share/nginx/html` - Not Started
   c. Configure `/api/` location to proxy to Core Service on port 8000 - Not Started
   d. Configure gzip compression for static assets - Not Started
   e. Set up caching headers for static files - Not Started
   f. Configure SPA routing to handle React Router paths - Not Started
   g. Copy built frontend files to nginx html directory in Dockerfile - Not Started

6. Build and test the Docker image - Not Started
   a. Create `.dockerignore` file to exclude unnecessary files from build context - Not Started
   b. Build the Docker image with `docker build -t chronocone-demo:latest` - Not Started
   c. Verify the image builds without errors - Not Started
   d. Run the container with `docker run -p 80:80 chronocone-demo:latest` - Not Started
   e. Verify container starts without errors in logs - Not Started
   f. Verify PostgreSQL is running and accessible within the container - Not Started
   g. Verify database tables are created successfully - Not Started
   h. Verify definitions are loaded from JSON file - Not Started
   i. Verify Core Service is running and responds to health checks - Not Started
   j. Verify frontend is served on port 80 and loads in browser - Not Started
   k. Verify API calls from frontend are proxied correctly to backend - Not Started

7. Create Docker Compose configuration (optional enhancement) - Not Started
   a. Create `docker-compose.yml` for easier container management - Not Started
   b. Configure volume for persistent PostgreSQL data - Not Started
   c. Configure volume for custom definitions JSON file - Not Started
   d. Add environment variable configuration for customization - Not Started
   e. Document usage in README - Not Started

8. Document the demo container - Not Started
   a. Create `docker/README.md` with container documentation - Not Started
   b. Document prerequisites: Docker installed, ports 80 available - Not Started
   c. Document how to build the container - Not Started
   d. Document how to run the container with various options - Not Started
   e. Document how to customize the definitions JSON file - Not Started
   f. Document how to access the application (URLs, default credentials if any) - Not Started
   g. Document troubleshooting steps for common issues - Not Started
   h. Document how to stop and remove the container - Not Started

9. Create PDD script for building and running the container - Not Started
   a. Create `scripts/build-demo-container.pdd.script.md` - Not Started
   b. Define script parameters: tag, build_args - Not Started
   c. Document prerequisites: Docker running, frontend built - Not Started
   d. Define operations: build image, verify build, tag image - Not Started
   e. Include verification steps to confirm successful build - Not Started
   f. Create `scripts/run-demo-container.pdd.script.md` - Not Started
   g. Define run parameters: port mapping, volume mounts, environment variables - Not Started
   h. Define verification steps to confirm container is running correctly - Not Started

10. Test end-to-end demonstration workflow - Not Started
    a. Build the container from scratch - Not Started
    b. Start the container and wait for services to initialize - Not Started
    c. Access the frontend in a browser and verify it loads - Not Started
    d. Navigate to Definitions page and verify demo definitions are displayed - Not Started
    e. Verify Core Service endpoint discovery works correctly - Not Started
    f. Verify Metrics Service is accessible through the Core Service proxy - Not Started
    g. Stop and restart the container to verify data persistence - Not Started
    h. Verify idempotency of definition loading (no duplicates created) - Not Started
    i. Document the complete demo workflow - Not Started

## Notes

- The container should be self-contained with no external dependencies except Docker
- PostgreSQL data should be persisted in a Docker volume so data survives container restarts
- The startup script should wait for PostgreSQL to be fully ready before attempting schema initialization
- Definition loading from JSON should be idempotent to allow container restarts without creating duplicates
- The JSON configuration file should be mounted as a volume so users can customize definitions without rebuilding the image
- Consider using supervisord or similar process manager to handle multiple services within the container
- The frontend should be built at Docker image build time, not at container runtime
- All services (PostgreSQL, Core Service, nginx) should be configured to log to stdout/stderr for easy debugging with `docker logs`
- The container should expose a health check endpoint that verifies all services are running
- The default demo definitions should provide a meaningful demonstration of ChronoCone capabilities
- Consider container size optimization: use distroless or alpine Go base images, clean up build artifacts, minimize layers
- Security considerations: do not include development credentials, use non-root users where possible, restrict PostgreSQL access to localhost
