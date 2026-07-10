import { X } from 'lucide-react';
import React, { useEffect, useRef, useState } from 'react';
import { TabulatorFull as Tabulator } from 'tabulator-tables';
import { attachMoreLinkListener, getCommonNameColumn } from '../../../lib/tabulatorUtils';
import { Category, CategoryElement, Taxonomy, TaxonomyNode } from '../../../types';

interface TaxonomiesProps {
  taxonomies: Taxonomy[];
  categories: Category[];
  expandedIds: string[];
  onToggleExpand: (id: string, expanded: boolean) => void;
  selectedId: string | null;
  onSelect: (id: string | null) => void;
  onAdd: () => void;
  onAddElement: (taxId: string, parentNodeId?: string) => void;
  onEdit: (id: string, tab?: 'general' | 'properties' | 'locales' | 'dag') => void;
  onGoToCategory: (catId: string, elId?: string) => void;
}

interface TaxonomyRow {
  id: string;
  key: string;
  name: string;
  tooltip: string;
  desc: string;
  isTax: boolean;
  _locales?: unknown[];
  _props?: unknown[];

  allowedCats?: string[];
  catId?: string;
  elId?: string;
  isMeasuresDimension?: boolean;
  _children?: TaxonomyRow[];
}

export const Taxonomies: React.FC<TaxonomiesProps> = ({
  taxonomies,
  categories,
  expandedIds,
  onToggleExpand,
  selectedId,
  onSelect,
  onAddElement,
  onEdit,
  onGoToCategory,
}) => {
  const tableRef = useRef<HTMLDivElement>(null);
  const tabulatorInstance = useRef<Tabulator | null>(null);
  const [isTableReady, setIsTableReady] = useState(false);
  const [overlayData, setOverlayData] = useState<{ id: string; name: string; desc: string } | null>(
    null
  );

  const isProgrammaticSelection = useRef(false);

  useEffect(() => {
    const findElement = (elKey: string): { cat: Category; el: CategoryElement } | null => {
      for (const cat of categories) {
        const el = cat._children.find(e => e.key === elKey);
        if (el) return { cat, el };
      }
      return null;
    };

    const mapDag = (nodes: TaxonomyNode[]): TaxonomyRow[] => {
      return nodes.map(node => {
        const found = findElement(node.elKey);
        const children = mapDag(node.children);
        const row: TaxonomyRow = {
          id: node.id,
          key: node.elKey,
          name: found?.el.name || node.elKey,
          tooltip: found?.el.tooltip || '',
          desc: found?.el.desc || '',
          isTax: false,
          catId: found?.cat.id,
          elId: found?.el.id,
          _locales: found?.el._locales || [],
        };

        if (children.length > 0) {
          row._children = children;
        }

        return row;
      });
    };

    if (tableRef.current) {
      // eslint-disable-next-line react-hooks/set-state-in-effect
      setIsTableReady(false);

      const tableData: TaxonomyRow[] = taxonomies.map(t => {
        const children = mapDag(t.dag);
        const row: TaxonomyRow = {
          id: t.id,
          key: t.key,
          name: t.name,
          tooltip: t.tooltip,
          desc: t.desc,
          allowedCats: t.allowedCats,
          _locales: t._locales,
          _props: t._props,
          isMeasuresDimension: t.isMeasuresDimension,
          isTax: true,
        };

        if (children.length > 0) {
          row._children = children;
        }

        return row;
      });

      tabulatorInstance.current = new Tabulator(tableRef.current, {
        data: tableData,
        dataTree: true,
        dataTreeChildField: '_children',
        dataTreeStartExpanded: (row: { getData: () => TaxonomyRow }) =>
          expandedIds.includes(row.getData().id),

        dataTreeChildIndent: 24,
        dataTreeElementColumn: 'name',
        layout: 'fitColumns',
        selectable: 1,
        selectableRangeMode: 'click',
        columns: [
          {
            formatter: 'rowSelection',
            titleFormatter: 'rowSelection',
            hozAlign: 'center',
            headerSort: false,
            width: 40,
            cellClick: (_e: unknown, cell: { getRow: () => { toggleSelect: () => void } }) => {
              cell.getRow().toggleSelect();
            },
          },
          {
            title: '',
            field: 'isMeasuresDimension',
            width: 32,
            hozAlign: 'center',
            headerHozAlign: 'center',
            headerSort: false,

            formatter: (c: {
              getRow: () => { getData: () => TaxonomyRow; getTreeParent: () => unknown };
            }) => {
              const d = c.getRow().getData() as TaxonomyRow;
              if (d.isMeasuresDimension) {
                const countNodes = (nodeList: TaxonomyRow[]): number => {
                  let count = 0;
                  for (const node of nodeList) {
                    count += 1;
                    if (node._children) {
                      count += countNodes(node._children);
                    }
                  }
                  return count;
                };
                const count = countNodes(d._children || []);

                return `
                    <div class="flex items-center justify-center min-w-[32px] h-[18px] px-1.5 rounded-full bg-[var(--surface2)] border border-[var(--border)] text-[10px] font-bold gap-1">
                      <span style="color: #3b82f6; font-size: 11px;">⎇</span>
                      <span style="color: var(--muted);">${count}</span>
                    </div>
                  `;
              }

              // Check if any parent is a measures dimension
              interface RowStub {
                getData: () => TaxonomyRow;
                getTreeParent: () => unknown;
              }
              let parent = c.getRow().getTreeParent() as RowStub | false;
              while (parent) {
                if (parent.getData().isMeasuresDimension) {
                  return `<span style="font-size: 14px;">🔹</span>`;
                }
                parent = parent.getTreeParent() as RowStub | false;
              }

              return '';
            },
          },
          getCommonNameColumn('Name', (id, name, desc) => {
            setOverlayData({ id, name, desc });
          }),
          {
            title: 'Categories',
            field: 'allowedCats',
            widthGrow: 1.5,
            formatter: (c: { getValue: () => string[] }) => {
              const cats = c.getValue() || [];
              return `<div class="flex flex-wrap gap-1">${cats.map((cat: string) => `<span class="px-2 py-0.5 bg-[var(--surface2)] border border-[var(--border)] rounded-full text-[10px] text-[var(--accent2)] font-medium">${cat}</span>`).join('')}</div>`;
            },
          },
          {
            title: 'Props',
            field: '_props',
            width: 70,
            hozAlign: 'center',
            headerHozAlign: 'center',
            formatter: (c: { getRow: () => { getData: () => TaxonomyRow } }) => {
              const d = c.getRow().getData();
              if (!d.isTax) return '';
              return `<button class="badge-btn badge-prop" data-taxid="${d.id}">⚙ ${(d._props || []).length}</button>`;
            },
          },
          {
            title: '',
            width: 80,
            hozAlign: 'center',
            headerSort: false,
            formatter: (c: { getRow: () => { getData: () => TaxonomyRow } }) => {
              const d = c.getRow().getData();
              if (d.isTax) {
                return `
                  <div class="flex items-center gap-1">
                    <button class="mc-edit" data-taxid="${d.id}" title="Edit">✎</button>
                    <button class="mc-add-el" data-taxid="${d.id}" title="Add Element">＋</button>
                  </div>
                `;
              } else {
                return `
                  <div class="flex items-center gap-1">
                    <button class="mc-add-el" data-nodeid="${d.id}" title="Add Child Element">＋</button>
                    <button class="mc-goto" data-catid="${d.catId}" data-elid="${d.elId}" title="Go to Categories">🔗</button>
                  </div>
                `;
              }
            },
          },
        ],
      });

      tabulatorInstance.current.on('dataTreeRowExpanded', (row: { getData: () => TaxonomyRow }) => {
        onToggleExpand(row.getData().id, true);
      });

      tabulatorInstance.current.on(
        'dataTreeRowCollapsed',
        (row: { getData: () => TaxonomyRow }) => {
          onToggleExpand(row.getData().id, false);
        }
      );

      tabulatorInstance.current.on('rowSelected', (row: { getData: () => TaxonomyRow }) => {
        if (!isProgrammaticSelection.current) {
          onSelect(row.getData().id);
        }
      });

      tabulatorInstance.current.on('rowDeselected', () => {
        if (!isProgrammaticSelection.current) {
          if (tabulatorInstance.current?.getSelectedRows().length === 0) {
            onSelect(null);
          }
        }
      });

      tabulatorInstance.current.on('tableBuilt', () => {
        setIsTableReady(true);
      });

      const tableEl = tableRef.current;
      const handleClick = (e: MouseEvent) => {
        const target = e.target as HTMLElement;
        const editBtn = target.closest('.mc-edit') as HTMLElement;
        const addElBtn = target.closest('.mc-add-el') as HTMLElement;
        const gotoBtn = target.closest('.mc-goto') as HTMLElement;
        const locBadge = target.closest('.badge-loc') as HTMLElement;
        const propBadge = target.closest('.badge-prop') as HTMLElement;

        if (editBtn) {
          const id = editBtn.dataset.taxid!;
          onEdit(id);
        } else if (locBadge) {
          const id = locBadge.dataset.taxid!;
          onEdit(id, 'locales');
        } else if (propBadge) {
          const id = propBadge.dataset.taxid!;
          onEdit(id, 'properties');
        } else if (addElBtn) {
          const taxId = addElBtn.dataset.taxid;
          const nodeId = addElBtn.dataset.nodeid;
          if (taxId) onAddElement(taxId);
          else if (nodeId) {
            const row = tabulatorInstance.current?.getRow(nodeId);
            let parent = row?.getTreeParent();
            while (parent && !parent.getData().isTax) {
              parent = parent.getTreeParent();
            }
            if (parent) {
              onAddElement(parent.getData().id, nodeId);
            }
          }
        } else if (gotoBtn) {
          const catId = gotoBtn.dataset.catid!;
          const elId = gotoBtn.dataset.elid!;
          onGoToCategory(catId, elId);
        }
      };

      tableEl.addEventListener('click', handleClick);

      // Attach the "More..." link listener
      const cleanup = attachMoreLinkListener(tableEl, (id, name, desc) => {
        setOverlayData({ id, name, desc });
      });

      return () => {
        tableEl.removeEventListener('click', handleClick);
        tabulatorInstance.current?.destroy();
        cleanup();
      };
    }
  }, [
    taxonomies,
    categories,
    onEdit,
    onAddElement,
    onGoToCategory,
    onToggleExpand,
    onSelect,
    expandedIds,
  ]);

  useEffect(() => {
    if (isTableReady && tabulatorInstance.current) {
      const currentSelected = tabulatorInstance.current.getSelectedRows();
      const currentId = currentSelected.length > 0 ? currentSelected[0].getData().id : null;

      if (selectedId !== currentId) {
        isProgrammaticSelection.current = true;
        if (selectedId) {
          const row = tabulatorInstance.current.getRow(selectedId);
          if (row) {
            const parent = row.getTreeParent();
            if (parent) parent.treeExpand();
            tabulatorInstance.current.selectRow(selectedId);
            row.scrollTo();
          } else {
            tabulatorInstance.current.deselectRow();
          }
        } else {
          tabulatorInstance.current.deselectRow();
        }
        isProgrammaticSelection.current = false;
      }
    }
  }, [isTableReady, selectedId]);

  return (
    <div className="tax-view h-full flex flex-col animate-in fade-in duration-300 relative">
      <div ref={tableRef} id="tax-tabulator" className="flex-1 overflow-hidden"></div>

      {/* Description Overlay */}
      {overlayData && (
        <div className="fixed inset-0 z-[2000] flex items-center justify-center bg-black/40 backdrop-blur-[2px] animate-in fade-in duration-200">
          <div className="bg-[var(--surface)] border border-[var(--border)] rounded-xl shadow-2xl w-full max-w-lg overflow-hidden flex flex-col animate-in zoom-in-95 duration-200">
            <div className="p-4 border-b border-[var(--border)] flex items-center justify-between bg-[var(--surface2)]/50">
              <h3 className="font-semibold text-[var(--text)]">{overlayData.name}</h3>
              <button
                onClick={() => setOverlayData(null)}
                className="p-1 rounded-lg hover:bg-[var(--surface2)] text-[var(--muted)] transition-colors"
              >
                <X size={18} />
              </button>
            </div>
            <div className="p-6 overflow-y-auto max-h-[60vh]">
              <p className="text-[14px] text-[var(--text)] leading-relaxed whitespace-pre-wrap">
                {overlayData.desc}
              </p>
            </div>
            <div className="p-4 border-t border-[var(--border)] flex justify-end bg-[var(--surface2)]/30">
              <button
                onClick={() => setOverlayData(null)}
                className="px-4 py-1.5 bg-[var(--surface2)] border border-[var(--border)] rounded-lg text-sm font-medium hover:bg-[var(--border)] transition-all"
              >
                Close
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
