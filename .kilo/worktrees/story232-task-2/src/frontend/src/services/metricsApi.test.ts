import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import {
  fetchMetadataTypes,
  fetchMetadataList,
  createMetadata,
  toBackendType,
  toFrontendType,
  isSupportedType,
  getMetadataTypeDisplayInfo,
  sortTypesForDisplay,
  ApiError,
  isApiError,
} from './metricsApi';
// eslint-disable-next-line @typescript-eslint/no-unused-vars
import * as _config from '../config';

// Mock the config module
vi.mock('../config', () => ({
  getMetricsApiUrl: vi.fn((path: string) => `http://localhost:8001${path}`),
}));

describe('metricsApi', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('fetchMetadataTypes', () => {
    it('fetches metadata types successfully', async () => {
      const mockTypes = ['namespace', 'category', 'element'];
      global.fetch = vi.fn().mockResolvedValueOnce({
        ok: true,
        json: async () => ({ types: mockTypes }),
      });

      const result = await fetchMetadataTypes();

      expect(result).toEqual(mockTypes);
      expect(fetch).toHaveBeenCalledWith('http://localhost:8001/types');
    });

    it('throws ApiError on failed response', async () => {
      global.fetch = vi.fn().mockResolvedValue({
        ok: false,
        status: 500,
        text: async () => 'Server error',
      });

      await expect(fetchMetadataTypes()).rejects.toThrow(ApiError);
      await expect(fetchMetadataTypes()).rejects.toThrow('Failed to fetch metadata types');
    });

    it('throws network error on fetch failure', async () => {
      global.fetch = vi.fn().mockRejectedValueOnce(new Error('Network error'));

      await expect(fetchMetadataTypes()).rejects.toThrow(ApiError);
      await expect(fetchMetadataTypes()).rejects.toThrow(
        'Network error while fetching metadata types'
      );
    });
  });

  describe('fetchMetadataList', () => {
    it('fetches all metadata items', async () => {
      const mockItems = [
        { type: 'namespace', name: 'Production', description: 'Prod env' },
        { type: 'category', name: 'Revenue', description: 'Revenue metrics' },
      ];
      global.fetch = vi.fn().mockResolvedValueOnce({
        ok: true,
        json: async () => ({ items: mockItems }),
      });

      const result = await fetchMetadataList();

      expect(result).toEqual(mockItems);
      expect(fetch).toHaveBeenCalledWith('http://localhost:8001/list');
    });

    it('fetches metadata items filtered by type', async () => {
      const mockItems = [{ type: 'namespace', name: 'Production', description: 'Prod env' }];
      global.fetch = vi.fn().mockResolvedValueOnce({
        ok: true,
        json: async () => ({ items: mockItems }),
      });

      const result = await fetchMetadataList('namespace');

      expect(result).toEqual(mockItems);
      expect(fetch).toHaveBeenCalledWith('http://localhost:8001/list?metadata_type=namespace');
    });

    it('throws ApiError on failed response', async () => {
      global.fetch = vi.fn().mockResolvedValueOnce({
        ok: false,
        text: async () => 'Not found',
      });

      await expect(fetchMetadataList()).rejects.toThrow(ApiError);
    });
  });

  describe('createMetadata', () => {
    it('creates metadata successfully', async () => {
      const mockResponse = { id: '123', message: 'Created successfully' };
      global.fetch = vi.fn().mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse,
      });

      const result = await createMetadata('namespace', 'Test', 'Test description', 'Test tooltip');

      expect(result).toEqual(mockResponse);
      expect(fetch).toHaveBeenCalledWith('http://localhost:8001/metadata', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          metadata_type: 'namespace',
          name: 'Test',
          description: 'Test description',
          tooltip: 'Test tooltip',
        }),
      });
    });

    it('throws ApiError on failed response', async () => {
      global.fetch = vi.fn().mockResolvedValueOnce({
        ok: false,
        text: async () => 'Validation error',
      });

      await expect(createMetadata('namespace', 'Test', 'Description', 'Tooltip')).rejects.toThrow(
        ApiError
      );
    });
  });

  describe('type mapping', () => {
    describe('toBackendType', () => {
      it('maps frontend type to backend type', () => {
        expect(toBackendType('namespace')).toBe('namespace');
        expect(toBackendType('namespace-family')).toBe('namespace_family');
        expect(toBackendType('category')).toBe('category');
        expect(toBackendType('element')).toBe('element');
        expect(toBackendType('document-type')).toBe('ts_doc_type');
      });

      it('returns null for unsupported types', () => {
        expect(toBackendType('namespace-type')).toBeNull();
        expect(toBackendType('document-row-type')).toBeNull();
      });

      it('returns null for unknown types', () => {
        expect(toBackendType('unknown-type')).toBeNull();
      });
    });

    describe('toFrontendType', () => {
      it('maps backend type to frontend type', () => {
        expect(toFrontendType('namespace')).toBe('namespace');
        expect(toFrontendType('namespace_family')).toBe('namespace-family');
        expect(toFrontendType('category')).toBe('category');
        expect(toFrontendType('element')).toBe('element');
        expect(toFrontendType('ts_doc_type')).toBe('document-type');
      });

      it('returns backend type for unknown types', () => {
        expect(toFrontendType('unknown_type')).toBe('unknown_type');
      });
    });

    describe('isSupportedType', () => {
      it('returns true for supported frontend types', () => {
        expect(isSupportedType('namespace')).toBe(true);
        expect(isSupportedType('namespace-family')).toBe(true);
        expect(isSupportedType('category')).toBe(true);
        expect(isSupportedType('element')).toBe(true);
        expect(isSupportedType('document-type')).toBe(true);
      });

      it('returns false for unsupported types', () => {
        expect(isSupportedType('namespace-type')).toBe(false);
        expect(isSupportedType('document-row-type')).toBe(false);
      });
    });

    describe('getMetadataTypeDisplayInfo', () => {
      it('returns display info for known types', () => {
        const nsInfo = getMetadataTypeDisplayInfo('namespace');
        expect(nsInfo.name).toBe('Namespaces');
        expect(nsInfo.description).toContain('Organizational units');
      });

      it('returns default info for unknown types', () => {
        const info = getMetadataTypeDisplayInfo('unknown');
        expect(info.name).toBe('unknown');
        expect(info.description).toBe('');
      });
    });

    describe('sortTypesForDisplay', () => {
      it('sorts types in display order', () => {
        const input = ['element', 'namespace', 'category'];
        const result = sortTypesForDisplay(input);

        expect(result.indexOf('namespace')).toBeLessThan(result.indexOf('category'));
        expect(result.indexOf('category')).toBeLessThan(result.indexOf('element'));
      });

      it('filters out types not in display order', () => {
        const input = ['namespace', 'unknown_type', 'category'];
        const result = sortTypesForDisplay(input);

        expect(result).toContain('namespace');
        expect(result).toContain('category');
        expect(result).not.toContain('unknown_type');
      });
    });
  });

  describe('ApiError', () => {
    it('creates error with correct properties', () => {
      const error = new ApiError('Test error', 500, true);

      expect(error.message).toBe('Test error');
      expect(error.statusCode).toBe(500);
      expect(error.isNetworkError).toBe(true);
      expect(error.name).toBe('ApiError');
    });

    it('creates error with default values', () => {
      const error = new ApiError('Test error');

      expect(error.message).toBe('Test error');
      expect(error.statusCode).toBeUndefined();
      expect(error.isNetworkError).toBe(false);
    });
  });

  describe('isApiError', () => {
    it('returns true for ApiError instances', () => {
      const error = new ApiError('Test');
      expect(isApiError(error)).toBe(true);
    });

    it('returns false for regular errors', () => {
      expect(isApiError(new Error('Test'))).toBe(false);
    });

    it('returns false for non-errors', () => {
      expect(isApiError('string')).toBe(false);
      expect(isApiError(null)).toBe(false);
      expect(isApiError(undefined)).toBe(false);
    });
  });
});
