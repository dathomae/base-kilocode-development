import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { MetadataTable, type MetadataType } from './MetadataTable';

describe('MetadataTable component', () => {
  const mockMetadataType: MetadataType = {
    id: 'test-type',
    name: 'Test Type',
    description: 'A test metadata type',
    items: [
      { id: '1', name: 'Item 1', description: 'First item' },
      { id: '2', name: 'Item 2', description: 'Second item' },
    ],
  };

  const emptyMetadataType: MetadataType = {
    id: 'empty-type',
    name: 'Empty Type',
    description: 'An empty metadata type',
    items: [],
  };

  it('renders the metadata table container', () => {
    render(<MetadataTable metadataType={mockMetadataType} />);

    const table = screen.getByTestId('metadata-table-test-type');
    expect(table).toBeInTheDocument();
  });

  it('displays the metadata type name', () => {
    render(<MetadataTable metadataType={mockMetadataType} />);

    const title = screen.getByRole('heading', { level: 3, name: 'Test Type' });
    expect(title).toBeInTheDocument();
  });

  it('displays the metadata type description', () => {
    render(<MetadataTable metadataType={mockMetadataType} />);

    const description = screen.getByText('A test metadata type');
    expect(description).toBeInTheDocument();
  });

  it('renders the add button', () => {
    render(<MetadataTable metadataType={mockMetadataType} />);

    const addButton = screen.getByTestId('add-btn-test-type');
    expect(addButton).toBeInTheDocument();
  });

  it('displays table headers correctly', () => {
    render(<MetadataTable metadataType={mockMetadataType} />);

    const nameHeader = screen.getByText('Name');
    const descriptionHeader = screen.getByText('Description');

    expect(nameHeader).toBeInTheDocument();
    expect(descriptionHeader).toBeInTheDocument();
  });

  it('renders all items in the table', () => {
    render(<MetadataTable metadataType={mockMetadataType} />);

    const item1 = screen.getByText('Item 1');
    const item2 = screen.getByText('Item 2');

    expect(item1).toBeInTheDocument();
    expect(item2).toBeInTheDocument();
  });

  it('renders item descriptions', () => {
    render(<MetadataTable metadataType={mockMetadataType} />);

    const desc1 = screen.getByText('First item');
    const desc2 = screen.getByText('Second item');

    expect(desc1).toBeInTheDocument();
    expect(desc2).toBeInTheDocument();
  });

  it('shows input fields when add button is clicked', () => {
    render(<MetadataTable metadataType={mockMetadataType} />);

    const addButton = screen.getByTestId('add-btn-test-type');
    fireEvent.click(addButton);

    const nameInput = screen.getByTestId('input-name-test-type');
    const descriptionInput = screen.getByTestId('input-description-test-type');

    expect(nameInput).toBeInTheDocument();
    expect(descriptionInput).toBeInTheDocument();
  });

  it('displays empty state message when no items exist', () => {
    render(<MetadataTable metadataType={emptyMetadataType} />);

    const emptyMessage = screen.getByTestId('empty-state-empty-type');
    expect(emptyMessage).toBeInTheDocument();
    expect(emptyMessage).toHaveTextContent(/no items yet/i);
  });

  it('calls onAddItem with correct data when save is clicked', () => {
    const mockOnAddItem = vi.fn();
    render(<MetadataTable metadataType={mockMetadataType} onAddItem={mockOnAddItem} />);

    const addButton = screen.getByTestId('add-btn-test-type');
    fireEvent.click(addButton);

    const nameInput = screen.getByTestId('input-name-test-type');
    const descriptionInput = screen.getByTestId('input-description-test-type');

    fireEvent.change(nameInput, { target: { value: 'New Item' } });
    fireEvent.change(descriptionInput, { target: { value: 'New description' } });

    const saveButton = screen.getByTestId('save-btn-test-type');
    fireEvent.click(saveButton);

    expect(mockOnAddItem).toHaveBeenCalledWith('test-type', {
      name: 'New Item',
      description: 'New description',
    });
  });

  it('does not call onAddItem when save is clicked with empty name', () => {
    const mockOnAddItem = vi.fn();
    render(<MetadataTable metadataType={mockMetadataType} onAddItem={mockOnAddItem} />);

    const addButton = screen.getByTestId('add-btn-test-type');
    fireEvent.click(addButton);

    const saveButton = screen.getByTestId('save-btn-test-type');
    fireEvent.click(saveButton);

    expect(mockOnAddItem).not.toHaveBeenCalled();
  });

  it('hides input fields and shows add button after save', () => {
    const mockOnAddItem = vi.fn();
    render(<MetadataTable metadataType={mockMetadataType} onAddItem={mockOnAddItem} />);

    const addButton = screen.getByTestId('add-btn-test-type');
    fireEvent.click(addButton);

    const nameInput = screen.getByTestId('input-name-test-type');
    fireEvent.change(nameInput, { target: { value: 'Test' } });

    const saveButton = screen.getByTestId('save-btn-test-type');
    fireEvent.click(saveButton);

    expect(screen.queryByTestId('input-name-test-type')).not.toBeInTheDocument();
    expect(screen.getByTestId('add-btn-test-type')).toBeInTheDocument();
  });

  it('cancels editing when cancel button is clicked', () => {
    render(<MetadataTable metadataType={mockMetadataType} />);

    const addButton = screen.getByTestId('add-btn-test-type');
    fireEvent.click(addButton);

    const nameInput = screen.getByTestId('input-name-test-type');
    fireEvent.change(nameInput, { target: { value: 'Cancelled Item' } });

    const cancelButton = screen.getByTestId('cancel-btn-test-type');
    fireEvent.click(cancelButton);

    expect(screen.queryByTestId('input-name-test-type')).not.toBeInTheDocument();
    expect(screen.queryByText('Cancelled Item')).not.toBeInTheDocument();
  });

  it('saves on Enter key press', () => {
    const mockOnAddItem = vi.fn();
    render(<MetadataTable metadataType={mockMetadataType} onAddItem={mockOnAddItem} />);

    const addButton = screen.getByTestId('add-btn-test-type');
    fireEvent.click(addButton);

    const nameInput = screen.getByTestId('input-name-test-type');
    fireEvent.change(nameInput, { target: { value: 'Enter Test' } });
    fireEvent.keyDown(nameInput, { key: 'Enter' });

    expect(mockOnAddItem).toHaveBeenCalledWith('test-type', {
      name: 'Enter Test',
      description: '',
    });
  });

  it('cancels on Escape key press', () => {
    render(<MetadataTable metadataType={mockMetadataType} />);

    const addButton = screen.getByTestId('add-btn-test-type');
    fireEvent.click(addButton);

    const nameInput = screen.getByTestId('input-name-test-type');
    fireEvent.change(nameInput, { target: { value: 'Escape Test' } });
    fireEvent.keyDown(nameInput, { key: 'Escape' });

    expect(screen.queryByTestId('input-name-test-type')).not.toBeInTheDocument();
  });

  it('disables add button while editing', () => {
    render(<MetadataTable metadataType={mockMetadataType} />);

    const addButton = screen.getByTestId('add-btn-test-type') as HTMLButtonElement;
    fireEvent.click(addButton);

    expect(addButton.disabled).toBe(true);
  });

  it('trims whitespace from input values', () => {
    const mockOnAddItem = vi.fn();
    render(<MetadataTable metadataType={mockMetadataType} onAddItem={mockOnAddItem} />);

    const addButton = screen.getByTestId('add-btn-test-type');
    fireEvent.click(addButton);

    const nameInput = screen.getByTestId('input-name-test-type');
    fireEvent.change(nameInput, { target: { value: '  Trimmed  ' } });

    const saveButton = screen.getByTestId('save-btn-test-type');
    fireEvent.click(saveButton);

    expect(mockOnAddItem).toHaveBeenCalledWith('test-type', {
      name: 'Trimmed',
      description: '',
    });
  });

  it('renders without onAddItem callback (view-only mode)', () => {
    render(<MetadataTable metadataType={mockMetadataType} />);

    const addButton = screen.getByTestId('add-btn-test-type');
    fireEvent.click(addButton);

    const nameInput = screen.getByTestId('input-name-test-type');
    fireEvent.change(nameInput, { target: { value: 'Test' } });

    const saveButton = screen.getByTestId('save-btn-test-type');
    fireEvent.click(saveButton);

    // Without onAddItem, the item should not be saved and inputs remain visible
    // since there's no callback to complete the save operation
    expect(screen.getByTestId('input-name-test-type')).toBeInTheDocument();
  });

  it('has correct ARIA label on add button', () => {
    render(<MetadataTable metadataType={mockMetadataType} />);

    const addButton = screen.getByLabelText(/add new test type/i);
    expect(addButton).toBeInTheDocument();
  });

  describe('isSaving prop', () => {
    it('disables add button when isSaving is true', () => {
      render(<MetadataTable metadataType={mockMetadataType} isSaving={true} />);

      const addButton = screen.getByTestId('add-btn-test-type') as HTMLButtonElement;
      expect(addButton.disabled).toBe(true);
    });

    it('enables add button when isSaving is false', () => {
      render(<MetadataTable metadataType={mockMetadataType} isSaving={false} />);

      const addButton = screen.getByTestId('add-btn-test-type') as HTMLButtonElement;
      expect(addButton.disabled).toBe(false);
    });

    it('disables add button when isAdding is true regardless of isSaving', () => {
      render(<MetadataTable metadataType={mockMetadataType} isSaving={false} />);

      const addButton = screen.getByTestId('add-btn-test-type') as HTMLButtonElement;
      fireEvent.click(addButton);

      expect(addButton.disabled).toBe(true);
    });

    it('add button is enabled by default when no saving state', () => {
      render(<MetadataTable metadataType={mockMetadataType} />);

      const addButton = screen.getByTestId('add-btn-test-type') as HTMLButtonElement;
      expect(addButton.disabled).toBe(false);
    });
  });
});
