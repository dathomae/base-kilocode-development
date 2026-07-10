import { PanelLeftClose, PanelLeftOpen } from 'lucide-react';
import React, { useState } from 'react';
import { cn } from '../../lib/utils';
import { TreeNode } from '../../types';

interface SidebarProps {
  tree: TreeNode[];
  onSelect: (node: TreeNode) => void;
  selectedNodeId?: number;
  isCollapsed: boolean;
  onToggleCollapse: () => void;
}

const TreeRow: React.FC<{
  node: TreeNode;
  depth: number;
  onSelect: (node: TreeNode) => void;
  selectedNodeId?: number;
  isCollapsed: boolean;
}> = ({ node, depth, onSelect, selectedNodeId, isCollapsed }) => {
  const isAlwaysExpanded = node.type === 'taxonomy' || node.type === 'measures';
  const [isExpanded, setIsExpanded] = useState(true);
  const hasChildren = node.children && node.children.length > 0;
  const showTriangle = hasChildren && !isAlwaysExpanded && !isCollapsed;

  return (
    <div className="tree-node">
      <div
        className={cn(
          'flex items-center gap-1 py-1.5 px-2 cursor-pointer border-b border-white/3 text-[13px] hover:bg-[#6c63ff]/12 transition-all',
          selectedNodeId === node.id && 'bg-[#6c63ff]/22 border-l-3 border-[var(--accent)]',
          isCollapsed && 'justify-center px-0'
        )}
        style={{ paddingLeft: isCollapsed ? '0' : `${8 + depth * 18}px` }}
        onClick={() => onSelect(node)}
        title={isCollapsed ? node.label : undefined}
      >
        {!isCollapsed &&
          (showTriangle ? (
            <button
              className="w-4 h-4 shrink-0 bg-none border-none cursor-pointer text-[10px] text-[var(--accent)] p-0 flex items-center justify-center"
              onClick={e => {
                e.stopPropagation();
                setIsExpanded(!isExpanded);
              }}
            >
              {isExpanded ? '▼' : '▶'}
            </button>
          ) : (
            <span className="w-4 h-4 shrink-0" />
          ))}
        <span
          className={cn(
            'shrink-0 flex items-center justify-center',
            isCollapsed ? 'text-[18px]' : ''
          )}
        >
          {node.icon}
        </span>
        {!isCollapsed && (
          <span className="overflow-hidden text-ellipsis whitespace-nowrap ml-1">{node.label}</span>
        )}
      </div>
      {hasChildren && (isExpanded || isAlwaysExpanded) && !isCollapsed && (
        <div className="tree-children">
          {node.children!.map(child => (
            <TreeRow
              key={child.id}
              node={child}
              depth={depth + 1}
              onSelect={onSelect}
              selectedNodeId={selectedNodeId}
              isCollapsed={isCollapsed}
            />
          ))}
        </div>
      )}
    </div>
  );
};

export const Sidebar: React.FC<SidebarProps> = ({
  tree,
  onSelect,
  selectedNodeId,
  isCollapsed,
  onToggleCollapse,
}) => {
  const findPath = (
    nodes: TreeNode[],
    targetId: number,
    path: TreeNode[] = []
  ): TreeNode[] | null => {
    for (const node of nodes) {
      const currentPath = [...path, node];
      if (node.id === targetId) return currentPath;
      if (node.children) {
        const result = findPath(node.children, targetId, currentPath);
        if (result) return result;
      }
    }
    return null;
  };

  const selectedPath = selectedNodeId ? findPath(tree, selectedNodeId) : null;

  return (
    <div
      id="tree-panel"
      className={cn(
        'shrink-0 bg-[var(--surface)] border-r border-[var(--border)] flex flex-col overflow-hidden transition-all duration-300 ease-in-out',
        isCollapsed ? 'w-[56px]' : 'w-[240px]'
      )}
    >
      <div
        className={cn(
          'p-2.5 px-3.5 flex items-center justify-between border-b border-[var(--border)] shrink-0 h-[37px]',
          isCollapsed && 'px-0 justify-center'
        )}
      >
        {!isCollapsed && (
          <span className="text-[11px] font-bold uppercase tracking-[0.8px] text-[var(--muted)]">
            Navigation
          </span>
        )}
        <button
          onClick={onToggleCollapse}
          className="p-1 rounded hover:bg-[var(--surface2)] text-[var(--muted)] hover:text-[var(--text)] transition-colors"
          title={isCollapsed ? 'Expand Sidebar' : 'Collapse Sidebar'}
        >
          {isCollapsed ? <PanelLeftOpen size={16} /> : <PanelLeftClose size={16} />}
        </button>
      </div>
      <div className="flex-1 overflow-y-auto overflow-x-hidden scrollbar-thin scrollbar-thumb-[var(--border)]">
        {isCollapsed && selectedPath ? (
          <div className="flex flex-col items-center py-2 gap-2">
            {selectedPath.map((node, i) => (
              <React.Fragment key={node.id}>
                <div
                  className={cn(
                    'w-10 h-10 flex items-center justify-center rounded-lg cursor-pointer transition-all hover:bg-[#6c63ff]/12',
                    selectedNodeId === node.id
                      ? 'bg-[#6c63ff]/22 text-[var(--accent)] shadow-sm'
                      : 'text-[var(--muted)]'
                  )}
                  onClick={() => onSelect(node)}
                  title={node.label}
                >
                  <span className="text-[20px]">{node.icon}</span>
                </div>
                {i < selectedPath.length - 1 && (
                  <div className="w-px h-2 bg-[var(--border)] opacity-50" />
                )}
              </React.Fragment>
            ))}
          </div>
        ) : (
          tree.map(node => (
            <TreeRow
              key={node.id}
              node={node}
              depth={0}
              onSelect={onSelect}
              selectedNodeId={selectedNodeId}
              isCollapsed={isCollapsed}
            />
          ))
        )}
      </div>
    </div>
  );
};
