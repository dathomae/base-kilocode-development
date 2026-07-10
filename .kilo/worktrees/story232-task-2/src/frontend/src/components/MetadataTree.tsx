import { useState } from 'react';
import './MetadataTree.css';

export interface TreeNode {
  type: string;
  name: string;
  description: string;
  tooltip?: string;
  children: TreeNode[];
  // Parent names for uniquely identifying this node in the hierarchy
  ns_fam_name?: string;
  ns_name?: string;
  category_name?: string;
  element_name?: string;
}

export interface MetadataTreeProps {
  nodes: TreeNode[];
  onAddChild?: (
    parentType: string,
    parentNode: TreeNode | null,
    name: string,
    description: string,
    tooltip: string
  ) => void;
  isSaving?: boolean;
}

interface TreeNodeProps {
  node: TreeNode;
  level: number;
  onAddChild?: (
    parentType: string,
    parentNode: TreeNode | null,
    name: string,
    description: string,
    tooltip: string
  ) => void;
  isSaving?: boolean;
}

function TreeNodeComponent({ node, level, onAddChild, isSaving }: TreeNodeProps) {
  const [isExpanded, setIsExpanded] = useState(true);
  const [isAdding, setIsAdding] = useState(false);
  const [newName, setNewName] = useState('');
  const [newDescription, setNewDescription] = useState('');
  const [newTooltip, setNewTooltip] = useState('');

  const hasChildren = node.children.length > 0;
  const canAddChild =
    node.type === 'namespace_family' || node.type === 'namespace' || node.type === 'category';

  const getChildTypeLabel = (parentType: string): string => {
    switch (parentType) {
      case 'namespace_family':
        return 'Namespace';
      case 'namespace':
        return 'Category';
      case 'category':
        return 'Element';
      default:
        return 'Item';
    }
  };

  const getAddButtonText = (parentType: string): string => {
    switch (parentType) {
      case 'namespace_family':
        return 'Add Namespace';
      case 'namespace':
        return 'Add Category';
      case 'category':
        return 'Add Element';
      default:
        return 'Add';
    }
  };

  const handleToggle = () => {
    if (hasChildren) {
      setIsExpanded(!isExpanded);
    }
  };

  const handleAddClick = () => {
    setIsAdding(true);
  };

  const handleCancel = () => {
    setIsAdding(false);
    setNewName('');
    setNewDescription('');
    setNewTooltip('');
  };

  const handleSave = () => {
    if (newName.trim() && onAddChild) {
      onAddChild(
        node.type,
        node,
        newName.trim(),
        newDescription.trim(),
        newTooltip.trim() || newName.trim()
      );
      setNewName('');
      setNewDescription('');
      setNewTooltip('');
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

  const getNodeIcon = (type: string): string => {
    switch (type) {
      case 'namespace_family':
        return '🏢';
      case 'namespace':
        return '📁';
      case 'category':
        return '🏷️';
      case 'element':
        return '🔹';
      default:
        return '📄';
    }
  };

  const getNodeTypeLabel = (type: string): string => {
    switch (type) {
      case 'namespace_family':
        return 'Namespace Family';
      case 'namespace':
        return 'Namespace';
      case 'category':
        return 'Category';
      case 'element':
        return 'Element';
      default:
        return type;
    }
  };

  return (
    <div className="tree-node" style={{ marginLeft: level > 0 ? `${level * 24}px` : 0 }}>
      <div className="tree-node__content">
        <button
          className={`tree-node__toggle ${!hasChildren ? 'tree-node__toggle--leaf' : ''}`}
          onClick={handleToggle}
          disabled={!hasChildren}
          aria-label={isExpanded ? 'Collapse' : 'Expand'}
        >
          {hasChildren ? (isExpanded ? '▼' : '▶') : ' '}
        </button>

        <span className="tree-node__icon">{getNodeIcon(node.type)}</span>

        <div className="tree-node__info">
          <div className="tree-node__name-row">
            <span className="tree-node__name" title={node.tooltip}>
              {node.name}
            </span>
            {node.description && (
              <span className="tree-node__description" title={node.description}>
                {node.description}
              </span>
            )}
          </div>
          <span className="tree-node__type">{getNodeTypeLabel(node.type)}</span>
        </div>

        {canAddChild && (
          <button
            className="tree-node__add-btn"
            onClick={handleAddClick}
            disabled={isAdding || isSaving}
            aria-label={`Add ${getChildTypeLabel(node.type)}`}
          >
            {getAddButtonText(node.type)}
          </button>
        )}
      </div>

      {isAdding && (
        <div className="tree-node__form" style={{ marginLeft: `${(level + 1) * 24}px` }}>
          <input
            type="text"
            className="tree-node__input"
            placeholder={`${getChildTypeLabel(node.type)} name...`}
            value={newName}
            onChange={e => setNewName(e.target.value)}
            onKeyDown={handleKeyDown}
            autoFocus
          />
          <input
            type="text"
            className="tree-node__input"
            placeholder="Description..."
            value={newDescription}
            onChange={e => setNewDescription(e.target.value)}
            onKeyDown={handleKeyDown}
          />
          <input
            type="text"
            className="tree-node__input"
            placeholder="Tooltip (shown on hover)..."
            value={newTooltip}
            onChange={e => setNewTooltip(e.target.value)}
            onKeyDown={handleKeyDown}
          />
          <div className="tree-node__actions">
            <button className="tree-node__save-btn" onClick={handleSave}>
              Save
            </button>
            <button className="tree-node__cancel-btn" onClick={handleCancel}>
              Cancel
            </button>
          </div>
        </div>
      )}

      {hasChildren && isExpanded && (
        <div className="tree-node__children">
          {node.children.map(child => (
            <TreeNodeComponent
              key={`${child.type}-${child.name}`}
              node={child}
              level={level + 1}
              onAddChild={onAddChild}
              isSaving={isSaving}
            />
          ))}
        </div>
      )}
    </div>
  );
}

export function MetadataTree({ nodes, onAddChild, isSaving }: MetadataTreeProps) {
  const [isAddingRoot, setIsAddingRoot] = useState(false);
  const [newName, setNewName] = useState('');
  const [newDescription, setNewDescription] = useState('');
  const [newTooltip, setNewTooltip] = useState('');

  const handleAddRootClick = () => {
    setIsAddingRoot(true);
  };

  const handleCancelRoot = () => {
    setIsAddingRoot(false);
    setNewName('');
    setNewDescription('');
    setNewTooltip('');
  };

  const handleSaveRoot = () => {
    if (newName.trim() && onAddChild) {
      // Pass tooltip as a separate parameter (no longer in attributes)
      onAddChild(
        'namespace_family',
        null,
        newName.trim(),
        newDescription.trim(),
        newTooltip.trim() || newName.trim()
      );
      setNewName('');
      setNewDescription('');
      setNewTooltip('');
      setIsAddingRoot(false);
    }
  };

  const handleKeyDownRoot = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      handleSaveRoot();
    } else if (e.key === 'Escape') {
      handleCancelRoot();
    }
  };

  if (nodes.length === 0 && !isAddingRoot) {
    return (
      <div className="metadata-tree metadata-tree--empty">
        <p>No metadata defined yet.</p>
        <button
          className="tree-node__add-btn tree-node__add-btn--root"
          onClick={handleAddRootClick}
          disabled={isSaving}
        >
          Add Namespace Family
        </button>
      </div>
    );
  }

  return (
    <div className="metadata-tree">
      {!isAddingRoot ? (
        <div className="metadata-tree__root-action">
          <button
            className="tree-node__add-btn tree-node__add-btn--root"
            onClick={handleAddRootClick}
            disabled={isSaving}
          >
            Add Namespace Family
          </button>
        </div>
      ) : (
        <div className="tree-node__form tree-node__form--root">
          <input
            type="text"
            className="tree-node__input"
            placeholder="Namespace Family name..."
            value={newName}
            onChange={e => setNewName(e.target.value)}
            onKeyDown={handleKeyDownRoot}
            autoFocus
          />
          <input
            type="text"
            className="tree-node__input"
            placeholder="Description..."
            value={newDescription}
            onChange={e => setNewDescription(e.target.value)}
            onKeyDown={handleKeyDownRoot}
          />
          <input
            type="text"
            className="tree-node__input"
            placeholder="Tooltip (shown on hover)..."
            value={newTooltip}
            onChange={e => setNewTooltip(e.target.value)}
            onKeyDown={handleKeyDownRoot}
          />
          <div className="tree-node__actions">
            <button className="tree-node__save-btn" onClick={handleSaveRoot}>
              Save
            </button>
            <button className="tree-node__cancel-btn" onClick={handleCancelRoot}>
              Cancel
            </button>
          </div>
        </div>
      )}

      {nodes.map(node => (
        <TreeNodeComponent
          key={`${node.type}-${node.name}`}
          node={node}
          level={0}
          onAddChild={onAddChild}
          isSaving={isSaving}
        />
      ))}
    </div>
  );
}

export default MetadataTree;
