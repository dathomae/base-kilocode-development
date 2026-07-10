/**
 * Metrics API Service
 *
 * Provides functions to interact with the Metrics Service endpoints for
 * fetching metadata types, listing metadata items, and creating new metadata entries.
 */

import { getMetricsApiUrl } from '../config';

// ============================================================================
// Type Definitions
// ============================================================================

/**
 * Metadata item as returned by the API
 */
export interface ApiMetadataItem {
  type: string;
  name: string;
  description: string;
}

/**
 * Response from the GET /types endpoint
 */
export interface TypesResponse {
  types: string[];
}

/**
 * Response from the GET /list endpoint
 */
export interface ListResponse {
  items: ApiMetadataItem[];
}

/**
 * Request body for POST /metadata endpoint
 */
export interface MetadataCreateRequest {
  metadata_type: string;
  name: string;
  description: string;
  tooltip: string;
  attributes?: Record<string, unknown>;
}

/**
 * Hierarchy item from GET /list/hierarchy endpoint
 * Uses names only (no internal database IDs)
 */
export interface HierarchyItem {
  type: string;
  name: string;
  description: string;
  tooltip?: string;
  ns_fam_name?: string;
  ns_name?: string;
  category_name?: string;
  element_name?: string;
}

/**
 * Response from GET /list/hierarchy endpoint
 */
export interface HierarchyResponse {
  items: HierarchyItem[];
}

/**
 * Response from the POST /metadata endpoint
 */
export interface MetadataResponse {
  id: string;
  message: string;
}

// ============================================================================
// Error Handling
// ============================================================================

/**
 * Custom error class for API errors
 */
export class ApiError extends Error {
  statusCode?: number;
  isNetworkError: boolean;

  constructor(message: string, statusCode?: number, isNetworkError = false) {
    super(message);
    this.name = 'ApiError';
    this.statusCode = statusCode;
    this.isNetworkError = isNetworkError;
  }
}

/**
 * Type guard to check if an error is an ApiError
 */
export function isApiError(error: unknown): error is ApiError {
  return error instanceof ApiError;
}

// ============================================================================
// Logging Utility
// ============================================================================

/**
 * Log API calls in development mode
 */
function logApiCall(method: string, endpoint: string, data?: unknown): void {
  if (import.meta.env.DEV) {
    console.log(`[MetricsAPI] ${method} ${endpoint}`, data ? data : '');
  }
}

/**
 * Log API responses in development mode
 */
function logApiResponse(endpoint: string, response: unknown): void {
  if (import.meta.env.DEV) {
    console.log(`[MetricsAPI] Response from ${endpoint}:`, response);
  }
}

/**
 * Log API errors in development mode
 */
function logApiError(endpoint: string, error: unknown): void {
  if (import.meta.env.DEV) {
    console.error(`[MetricsAPI] Error from ${endpoint}:`, error);
  }
}

// ============================================================================
// API Functions
// ============================================================================

/**
 * Fetch all supported metadata types from the Metrics Service
 *
 * @returns Promise resolving to array of metadata type strings
 * @throws ApiError if the request fails
 */
export async function fetchMetadataTypes(): Promise<string[]> {
  const endpoint = '/types';
  logApiCall('GET', endpoint);

  try {
    const response = await fetch(getMetricsApiUrl(endpoint));

    if (!response.ok) {
      const errorText = await response.text().catch(() => 'Unknown error');
      logApiError(endpoint, { status: response.status, error: errorText });
      throw new ApiError(`Failed to fetch metadata types: ${errorText}`, response.status);
    }

    const data: TypesResponse = await response.json();
    logApiResponse(endpoint, data);
    return data.types;
  } catch (error) {
    if (error instanceof ApiError) {
      throw error;
    }

    // Network or other errors
    const message = error instanceof Error ? error.message : 'Unknown error';
    logApiError(endpoint, message);
    throw new ApiError(`Network error while fetching metadata types: ${message}`, undefined, true);
  }
}

/**
 * Fetch metadata items from the Metrics Service
 *
 * @param type Optional metadata type to filter by
 * @returns Promise resolving to array of metadata items
 * @throws ApiError if the request fails
 */
export async function fetchMetadataList(type?: string): Promise<ApiMetadataItem[]> {
  const endpoint = type ? `/list?metadata_type=${encodeURIComponent(type)}` : '/list';
  logApiCall('GET', endpoint);

  try {
    const response = await fetch(getMetricsApiUrl(endpoint));

    if (!response.ok) {
      const errorText = await response.text().catch(() => 'Unknown error');
      logApiError(endpoint, { status: response.status, error: errorText });
      throw new ApiError(`Failed to fetch metadata list: ${errorText}`, response.status);
    }

    const data: ListResponse = await response.json();
    logApiResponse(endpoint, data);
    return data.items;
  } catch (error) {
    if (error instanceof ApiError) {
      throw error;
    }

    // Network or other errors
    const message = error instanceof Error ? error.message : 'Unknown error';
    logApiError(endpoint, message);
    throw new ApiError(`Network error while fetching metadata list: ${message}`, undefined, true);
  }
}

/**
 * Fetch metadata hierarchy from the Metrics Service
 *
 * @returns Promise resolving to array of hierarchy items with parent relationships
 * @throws ApiError if the request fails
 */
export async function fetchMetadataHierarchy(): Promise<HierarchyItem[]> {
  const endpoint = '/list/hierarchy';
  logApiCall('GET', endpoint);

  try {
    const response = await fetch(getMetricsApiUrl(endpoint));

    if (!response.ok) {
      const errorText = await response.text().catch(() => 'Unknown error');
      logApiError(endpoint, { status: response.status, error: errorText });
      throw new ApiError(`Failed to fetch metadata hierarchy: ${errorText}`, response.status);
    }

    const data: HierarchyResponse = await response.json();
    logApiResponse(endpoint, data);
    return data.items;
  } catch (error) {
    if (error instanceof ApiError) {
      throw error;
    }

    // Network or other errors
    const message = error instanceof Error ? error.message : 'Unknown error';
    logApiError(endpoint, message);
    throw new ApiError(
      `Network error while fetching metadata hierarchy: ${message}`,
      undefined,
      true
    );
  }
}

/**
 * Create a new metadata entry in the Metrics Service
 *
 * @param metadataType The type of metadata to create
 * @param name The name of the metadata entry
 * @param description The description of the metadata entry
 * @param tooltip Short tooltip description (required for most metadata types)
 * @param attributes Additional attributes for the metadata (e.g., namespace_name for categories)
 * @returns Promise resolving to the created metadata response
 * @throws ApiError if the request fails
 */
export async function createMetadata(
  metadataType: string,
  name: string,
  description: string,
  tooltip: string,
  attributes?: Record<string, unknown>
): Promise<MetadataResponse> {
  const endpoint = '/metadata';
  const requestBody: MetadataCreateRequest = {
    metadata_type: metadataType,
    name,
    description,
    tooltip,
    attributes,
  };

  logApiCall('POST', endpoint, requestBody);

  try {
    const response = await fetch(getMetricsApiUrl(endpoint), {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(requestBody),
    });

    if (!response.ok) {
      const errorText = await response.text().catch(() => 'Unknown error');
      logApiError(endpoint, { status: response.status, error: errorText });
      throw new ApiError(`Failed to create metadata: ${errorText}`, response.status);
    }

    const data: MetadataResponse = await response.json();
    logApiResponse(endpoint, data);
    return data;
  } catch (error) {
    if (error instanceof ApiError) {
      throw error;
    }

    // Network or other errors
    const message = error instanceof Error ? error.message : 'Unknown error';
    logApiError(endpoint, message);
    throw new ApiError(`Network error while creating metadata: ${message}`, undefined, true);
  }
}

// ============================================================================
// Metadata Type Mapping
// ============================================================================

/**
 * Maps frontend metadata type IDs to backend metadata type names
 */
const FRONTEND_TO_BACKEND_TYPE_MAP: Record<string, string | null> = {
  namespace: 'namespace',
  'namespace-family': 'namespace_family',
  namespace_family: 'namespace_family',
  'namespace-type': null, // Not in backend - skip
  category: 'category',
  element: 'element',
  'document-type': 'ts_doc_type',
  'document-row-type': null, // Not in backend - skip
};

/**
 * Maps backend metadata type names to frontend metadata type IDs
 */
const BACKEND_TO_FRONTEND_TYPE_MAP: Record<string, string> = {
  namespace: 'namespace',
  namespace_family: 'namespace-family',
  category: 'category',
  element: 'element',
  ts_doc_type: 'document-type',
};

/**
 * UI display information for metadata types
 */
const METADATA_TYPE_DISPLAY_INFO: Record<string, { name: string; description: string }> = {
  namespace: {
    name: 'Namespaces',
    description: 'Organizational units for grouping related metrics and definitions',
  },
  'namespace-family': {
    name: 'Namespace Families',
    description: 'Collections of related namespaces for multi-tenant organization',
  },
  'namespace-type': {
    name: 'Namespace Types',
    description: 'Classification types for namespaces (Production, Test, etc.)',
  },
  category: {
    name: 'Categories',
    description: 'Taxonomy categories for organizing elements and metrics',
  },
  element: {
    name: 'Elements',
    description: 'Individual taxonomy elements within categories',
  },
  'document-type': {
    name: 'Document Types',
    description: 'Types of time-series documents for metric storage',
  },
  'document-row-type': {
    name: 'Document Row Types',
    description: 'Row types defining metric schema within documents',
  },
};

/**
 * Convert a frontend metadata type ID to a backend metadata type name
 *
 * @param frontendType The frontend type ID (e.g., 'namespace-family')
 * @returns The backend type name or null if not supported
 */
export function toBackendType(frontendType: string): string | null {
  return FRONTEND_TO_BACKEND_TYPE_MAP[frontendType] ?? null;
}

/**
 * Convert a backend metadata type name to a frontend metadata type ID
 *
 * @param backendType The backend type name (e.g., 'namespace_family')
 * @returns The frontend type ID
 */
export function toFrontendType(backendType: string): string {
  return BACKEND_TO_FRONTEND_TYPE_MAP[backendType] ?? backendType;
}

/**
 * Check if a frontend metadata type is supported by the backend
 *
 * @param frontendType The frontend type ID
 * @returns true if the type is supported
 */
export function isSupportedType(frontendType: string): boolean {
  return FRONTEND_TO_BACKEND_TYPE_MAP[frontendType] !== null;
}

/**
 * Get display information for a metadata type
 *
 * @param frontendType The frontend type ID
 * @returns Display name and description for the type
 */
export function getMetadataTypeDisplayInfo(frontendType: string): {
  name: string;
  description: string;
} {
  return (
    METADATA_TYPE_DISPLAY_INFO[frontendType] ?? {
      name: frontendType,
      description: '',
    }
  );
}

/**
 * Filter and sort backend types to match frontend display order
 *
 * @param backendTypes Array of backend type names
 * @returns Array of frontend type IDs in display order
 */
export function sortTypesForDisplay(backendTypes: string[]): string[] {
  const frontendTypesOrder = [
    'namespace',
    'namespace-family',
    'namespace-type',
    'category',
    'element',
    'document-type',
    'document-row-type',
  ];

  const availableTypes = new Set(backendTypes.map(toFrontendType));

  return frontendTypesOrder.filter(type => availableTypes.has(type));
}
