import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { Definitions } from './Definitions';
import * as metricsApi from '../services/metricsApi';

// Mock the metricsApi module
vi.mock('../services/metricsApi', () => ({
  fetchMetadataHierarchy: vi.fn(),
  createMetadata: vi.fn(),
  toBackendType: vi.fn((type: string) => {
    const map: Record<string, string | null> = {
      namespace: 'namespace',
      'namespace-family': 'namespace_family',
      namespace_family: 'namespace_family',
      category: 'category',
      element: 'element',
    };
    return map[type] || null;
  }),
  toFrontendType: vi.fn((type: string) => {
    const map: Record<string, string> = {
      namespace: 'namespace',
      namespace_family: 'namespace-family',
      category: 'category',
      element: 'element',
    };
    return map[type] || type;
  }),
  isSupportedType: vi.fn((type: string) => {
    const supported = ['namespace', 'namespace-family', 'category', 'element'];
    return supported.includes(type);
  }),
  getMetadataTypeDisplayInfo: vi.fn((type: string) => {
    const info: Record<string, { name: string; description: string }> = {
      namespace: {
        name: 'Namespaces',
        description: 'Organizational units for grouping related metrics',
      },
      'namespace-family': {
        name: 'Namespace Families',
        description: 'Collections of related namespaces',
      },
      category: {
        name: 'Categories',
        description: 'Taxonomy categories',
      },
      element: {
        name: 'Elements',
        description: 'Individual taxonomy elements',
      },
    };
    return info[type] || { name: type, description: '' };
  }),
  sortTypesForDisplay: vi.fn((types: string[]) => {
    const order = ['namespace', 'namespace-family', 'category', 'element'];
    return order.filter(t => types.includes(t));
  }),
  isApiError: vi.fn((error: unknown) => {
    return error instanceof Error && error.name === 'ApiError';
  }),
  ApiError: class ApiError extends Error {
    statusCode?: number;
    isNetworkError: boolean;

    constructor(message: string, statusCode?: number, isNetworkError = false) {
      super(message);
      this.name = 'ApiError';
      this.statusCode = statusCode;
      this.isNetworkError = isNetworkError;
    }
  },
}));

describe('Definitions page', () => {
  const mockFetchMetadataHierarchy = metricsApi.fetchMetadataHierarchy as ReturnType<typeof vi.fn>;
  const mockCreateMetadata = metricsApi.createMetadata as ReturnType<typeof vi.fn>;

  beforeEach(() => {
    vi.clearAllMocks();
    // Default successful response - empty hierarchy
    mockFetchMetadataHierarchy.mockResolvedValue([]);
  });

  it('renders the definitions page container', async () => {
    render(<Definitions />);

    await waitFor(() => {
      const definitionsPage = screen.getByTestId('definitions-page');
      expect(definitionsPage).toBeInTheDocument();
    });
  });

  it('displays the page title', async () => {
    render(<Definitions />);

    await waitFor(() => {
      const title = screen.getByRole('heading', { level: 1, name: /definitions/i });
      expect(title).toBeInTheDocument();
    });
  });

  it('displays the subtitle with page description', async () => {
    render(<Definitions />);

    await waitFor(() => {
      const subtitle = screen.getByText(/manage your metadata types/i);
      expect(subtitle).toBeInTheDocument();
    });
  });

  it('shows loading spinner while fetching data', async () => {
    // Use a delayed promise that never resolves during the test
    mockFetchMetadataHierarchy.mockImplementation(() => new Promise(() => {}));

    render(<Definitions />);

    // Check for loading text immediately after render
    const loadingSpinner = screen.getByText(/loading metadata/i);
    expect(loadingSpinner).toBeInTheDocument();
  });

  it('shows error message when API call fails', async () => {
    const { ApiError } = await import('../services/metricsApi');
    mockFetchMetadataHierarchy.mockRejectedValue(new ApiError('Network error', undefined, true));

    render(<Definitions />);

    await waitFor(() => {
      const errorMessage = screen.getByTestId('error-message');
      expect(errorMessage).toBeInTheDocument();
    });
  });

  it('shows retry button when error occurs', async () => {
    const { ApiError } = await import('../services/metricsApi');
    mockFetchMetadataHierarchy.mockRejectedValue(new ApiError('Network error'));

    render(<Definitions />);

    await waitFor(() => {
      const retryButton = screen.getByTestId('error-retry-button');
      expect(retryButton).toBeInTheDocument();
    });
  });

  it('calls fetchMetadataHierarchy on mount', async () => {
    render(<Definitions />);

    await waitFor(() => {
      expect(mockFetchMetadataHierarchy).toHaveBeenCalledTimes(1);
    });
  });

  it('displays add button for root namespace family when tree is empty', async () => {
    render(<Definitions />);

    await waitFor(() => {
      const addButton = screen.getByText('Add Namespace Family');
      expect(addButton).toBeInTheDocument();
    });
  });

  it('shows input fields when add button is clicked', async () => {
    render(<Definitions />);

    await waitFor(() => {
      screen.getByText('Add Namespace Family');
    });

    const addButton = screen.getByText('Add Namespace Family');
    fireEvent.click(addButton);

    const nameInput = screen.getByPlaceholderText(/namespace family name/i);
    expect(nameInput).toBeInTheDocument();
  });

  it('calls createMetadata when saving a new namespace family', async () => {
    mockCreateMetadata.mockResolvedValue({ id: '123', message: 'Created' });

    render(<Definitions />);

    await waitFor(() => {
      screen.getByText('Add Namespace Family');
    });

    const addButton = screen.getByText('Add Namespace Family');
    fireEvent.click(addButton);

    const nameInput = screen.getByPlaceholderText(/namespace family name/i);
    fireEvent.change(nameInput, { target: { value: 'New Family' } });

    const saveButton = screen.getByText('Save');
    fireEvent.click(saveButton);

    await waitFor(() => {
      expect(mockCreateMetadata).toHaveBeenCalledWith(
        'namespace_family',
        'New Family',
        '',
        'New Family',
        {}
      );
    });
  });

  it('refreshes data after successful save', async () => {
    mockCreateMetadata.mockResolvedValue({ id: '123', message: 'Created' });

    render(<Definitions />);

    await waitFor(() => {
      screen.getByText('Add Namespace Family');
    });

    const addButton = screen.getByText('Add Namespace Family');
    fireEvent.click(addButton);

    const nameInput = screen.getByPlaceholderText(/namespace family name/i);
    fireEvent.change(nameInput, { target: { value: 'New Family' } });

    const saveButton = screen.getByText('Save');
    fireEvent.click(saveButton);

    await waitFor(() => {
      // Should refetch hierarchy after save
      expect(mockFetchMetadataHierarchy).toHaveBeenCalledTimes(2);
    });
  });

  it('shows error when createMetadata fails', async () => {
    const { ApiError } = await import('../services/metricsApi');
    mockCreateMetadata.mockRejectedValue(new ApiError('Save failed'));

    render(<Definitions />);

    await waitFor(() => {
      screen.getByText('Add Namespace Family');
    });

    const addButton = screen.getByText('Add Namespace Family');
    fireEvent.click(addButton);

    const nameInput = screen.getByPlaceholderText(/namespace family name/i);
    fireEvent.change(nameInput, { target: { value: 'New Family' } });

    const saveButton = screen.getByText('Save');
    fireEvent.click(saveButton);

    await waitFor(() => {
      const errorMessage = screen.getByTestId('error-message');
      expect(errorMessage).toBeInTheDocument();
    });
  });

  it('has the correct CSS class on the container', async () => {
    render(<Definitions />);

    await waitFor(() => {
      const definitionsPage = screen.getByTestId('definitions-page');
      expect(definitionsPage).toHaveClass('definitions-page');
    });
  });

  it('displays intro text about browsing hierarchy', async () => {
    render(<Definitions />);

    await waitFor(() => {
      const introText = screen.getByText(/browse the hierarchy below/i);
      expect(introText).toBeInTheDocument();
    });
  });

  it('allows canceling the add operation', async () => {
    render(<Definitions />);

    await waitFor(() => {
      screen.getByText('Add Namespace Family');
    });

    const addButton = screen.getByText('Add Namespace Family');
    fireEvent.click(addButton);

    const cancelButton = screen.getByText('Cancel');
    fireEvent.click(cancelButton);

    // Form should be closed, button should be visible again
    await waitFor(() => {
      expect(screen.getByText('Add Namespace Family')).toBeInTheDocument();
    });
  });

  it('handles retry button click to reload data', async () => {
    const { ApiError } = await import('../services/metricsApi');
    mockFetchMetadataHierarchy
      .mockRejectedValueOnce(new ApiError('Network error'))
      .mockResolvedValueOnce([]);

    render(<Definitions />);

    await waitFor(() => {
      const retryButton = screen.getByTestId('error-retry-button');
      expect(retryButton).toBeInTheDocument();
    });

    const retryButton = screen.getByTestId('error-retry-button');
    fireEvent.click(retryButton);

    await waitFor(() => {
      expect(mockFetchMetadataHierarchy).toHaveBeenCalledTimes(2);
    });
  });

  it('shows empty state message when no metadata exists', async () => {
    render(<Definitions />);

    await waitFor(() => {
      const emptyMessage = screen.getByText('No metadata defined yet.');
      expect(emptyMessage).toBeInTheDocument();
    });
  });
});
