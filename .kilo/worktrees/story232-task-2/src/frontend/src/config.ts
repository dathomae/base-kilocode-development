/**
 * Centralized configuration module for environment variables
 *
 * This module provides typed access to environment variables used throughout
 * the CHRONOCONE frontend application.
 *
 * Environment variables must be prefixed with VITE_ to be exposed to client code.
 * See: https://vitejs.dev/guide/env-and-mode.html
 */

/**
 * Backend Core Server URL
 * Used for API calls to the backend core service
 * @default "http://localhost:8000"
 */
export const CORE_SERVER_URL: string =
  import.meta.env.VITE_CORE_SERVER_URL || 'http://localhost:8000';

/**
 * Metrics Service URL
 * Used for API calls to the metrics service
 * @default "http://localhost:8001"
 */
export const METRICS_SERVICE_URL: string =
  import.meta.env.VITE_METRICS_SERVICE_URL || 'http://localhost:8001';

/**
 * Configuration object containing all environment variables
 */
export const config = {
  coreServerUrl: CORE_SERVER_URL,
  metricsServiceUrl: METRICS_SERVICE_URL,
} as const;

/**
 * Helper function to get the full API URL by appending a path to the core server URL
 * @param path - The API path (e.g., "/api/v1/status")
 * @returns The full URL
 */
export function getApiUrl(path: string): string {
  const baseUrl = CORE_SERVER_URL.replace(/\/$/, ''); // Remove trailing slash
  const cleanPath = path.replace(/^\//, ''); // Remove leading slash
  return `${baseUrl}/${cleanPath}`;
}

/**
 * Helper function to get the full API URL by appending a path to the metrics service URL
 * @param path - The API path (e.g., "/types")
 * @returns The full URL
 */
export function getMetricsApiUrl(path: string): string {
  const baseUrl = METRICS_SERVICE_URL.replace(/\/$/, ''); // Remove trailing slash
  const cleanPath = path.replace(/^\//, ''); // Remove leading slash
  return `${baseUrl}/${cleanPath}`;
}
