import { useState } from 'react';
import './MetadataTable.css';

export interface MetadataItem {
  id: string;
  name: string;
  description: string;
}

export interface MetadataType {
  id: string;
  name: string;
  description: string;
  items: MetadataItem[];
}

export interface MetadataTableProps {
  metadataType: MetadataType;
  onAddItem?: (metadataTypeId: string, item: Omit<MetadataItem, 'id'>) => void;
  /**
   * Whether the component is in a saving state
   * When true, the add button is disabled
   * @default false
   */
  isSaving?: boolean;
}

export function MetadataTable({ metadataType, onAddItem, isSaving = false }: MetadataTableProps) {
  const [isAdding, setIsAdding] = useState(false);
  const [newItemName, setNewItemName] = useState('');
  const [newItemDescription, setNewItemDescription] = useState('');

  const handleAddClick = () => {
    setIsAdding(true);
  };

  const handleCancel = () => {
    setIsAdding(false);
    setNewItemName('');
    setNewItemDescription('');
  };

  const handleSave = () => {
    if (newItemName.trim() && onAddItem) {
      onAddItem(metadataType.id, {
        name: newItemName.trim(),
        description: newItemDescription.trim(),
      });
      setNewItemName('');
      setNewItemDescription('');
      setIsAdding(false);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      handleSave();
    } else if (e.key === 'Escape') {
      handleCancel();
    }
  };

  return (
    <div className="metadata-table" data-testid={`metadata-table-${metadataType.id}`}>
      <div className="metadata-table__header">
        <h3 className="metadata-table__title">{metadataType.name}</h3>
        <p className="metadata-table__description">{metadataType.description}</p>
        <button
          className="metadata-table__add-btn"
          onClick={handleAddClick}
          aria-label={`Add new ${metadataType.name}`}
          data-testid={`add-btn-${metadataType.id}`}
          disabled={isAdding || isSaving}
        >
          <span className="metadata-table__add-icon">+</span>
        </button>
      </div>

      <div className="metadata-table__container">
        <table className="metadata-table__table">
          <thead>
            <tr>
              <th className="metadata-table__th metadata-table__th--name">Name</th>
              <th className="metadata-table__th metadata-table__th--description">Description</th>
            </tr>
          </thead>
          <tbody>
            {metadataType.items.map(item => (
              <tr key={item.id} data-testid={`metadata-row-${item.id}`}>
                <td className="metadata-table__td metadata-table__td--name">{item.name}</td>
                <td className="metadata-table__td metadata-table__td--description">
                  {item.description}
                </td>
              </tr>
            ))}
            {isAdding && (
              <tr
                className="metadata-table__row--editing"
                data-testid={`editing-row-${metadataType.id}`}
              >
                <td className="metadata-table__td metadata-table__td--name">
                  <input
                    type="text"
                    className="metadata-table__input"
                    placeholder="Enter name..."
                    value={newItemName}
                    onChange={e => setNewItemName(e.target.value)}
                    onKeyDown={handleKeyDown}
                    autoFocus
                    data-testid={`input-name-${metadataType.id}`}
                  />
                </td>
                <td className="metadata-table__td metadata-table__td--description">
                  <input
                    type="text"
                    className="metadata-table__input"
                    placeholder="Enter description..."
                    value={newItemDescription}
                    onChange={e => setNewItemDescription(e.target.value)}
                    onKeyDown={handleKeyDown}
                    data-testid={`input-description-${metadataType.id}`}
                  />
                </td>
              </tr>
            )}
            {!isAdding && metadataType.items.length === 0 && (
              <tr>
                <td
                  colSpan={2}
                  className="metadata-table__td metadata-table__td--empty"
                  data-testid={`empty-state-${metadataType.id}`}
                >
                  No items yet. Click + to add a new {metadataType.name.toLowerCase()}.
                </td>
              </tr>
            )}
          </tbody>
        </table>
        {isAdding && (
          <div className="metadata-table__actions">
            <button
              className="metadata-table__save-btn"
              onClick={handleSave}
              data-testid={`save-btn-${metadataType.id}`}
            >
              Save
            </button>
            <button
              className="metadata-table__cancel-btn"
              onClick={handleCancel}
              data-testid={`cancel-btn-${metadataType.id}`}
            >
              Cancel
            </button>
          </div>
        )}
      </div>
    </div>
  );
}

export default MetadataTable;
