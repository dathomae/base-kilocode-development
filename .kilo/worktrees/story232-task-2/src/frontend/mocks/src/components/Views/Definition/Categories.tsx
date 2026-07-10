import { X } from 'lucide-react';
import React, { useEffect, useRef, useState } from 'react';
import { TabulatorFull as Tabulator } from 'tabulator-tables';
import { attachMoreLinkListener, getCommonNameColumn } from '../../../lib/tabulatorUtils';
import { Category } from '../../../types';

interface CategoryRow {
  id: string;
  key: string;
  name: string;
  tooltip: string;
  desc: string;
  isBaseUnit?: boolean;
  _locales?: unknown[];
  _props?: unknown[];

  _children?: CategoryRow[];
}

interface CategoriesProps {
  categories: Category[];
  expandedIds: string[];
  onToggleExpand: (id: string, expanded: boolean) => void;
  selectedId: string | null;
  onSelect: (id: string | null) => void;
  targetElId: string | null;
  onClearTarget: () => void;
  onAdd: () => void;
  onAddElement: (catId: string) => void;
  onEdit: (id: string, isCat: boolean, tab?: 'general' | 'properties' | 'locales') => void;
}

export const Categories: React.FC<CategoriesProps> = ({
  categories,
  expandedIds,
  onToggleExpand,
  selectedId,
  onSelect,
  targetElId,
  onClearTarget,
  onAddElement,
  onEdit,
}) => {
  const tableRef = useRef<HTMLDivElement>(null);
  const tabulatorInstance = useRef<Tabulator | null>(null);
  const [isTableReady, setIsTableReady] = useState(false);
  const [overlayData, setOverlayData] = useState<{ id: string; name: string; desc: string } | null>(
    null
  );

  const isProgrammaticSelection = useRef(false);

  useEffect(() => {
    if (tableRef.current) {
      setTimeout(() => setIsTableReady(false), 0);

      const tableData: CategoryRow[] = categories.map(cat => {
        const children: CategoryRow[] = (cat._children || []).map(el => ({
          id: el.id,
          key: el.key,
          name: el.name,
          tooltip: el.tooltip,
          desc: el.desc,
          isBaseUnit: el.isBaseUnit,
          _locales: el._locales,
          _props: el._props,
        }));

        const row: CategoryRow = {
          id: cat.id,
          key: cat.key,
          name: cat.name,
          tooltip: cat.tooltip,
          desc: cat.desc,
          isBaseUnit: cat.isBaseUnit,
          _locales: cat._locales,
          _props: cat._props,
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
        dataTreeStartExpanded: (row: { getData: () => CategoryRow }) =>
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
            field: 'isBaseUnit',
            width: 32,
            hozAlign: 'center',
            headerHozAlign: 'center',
            headerSort: false,

            formatter: (c: {
              getRow: () => { getData: () => CategoryRow; getTreeParent: () => unknown };
              getValue: () => boolean;
            }) => {
              const d = c.getRow().getData() as CategoryRow;
              const isCat = !!d._children;
              const isUnitCat = d.key === 'channel' || d.key === 'product_class';

              if (isCat && isUnitCat) {
                const count = (d._children || []).length;
                return `
                    <div class="flex items-center justify-center min-w-[32px] h-[18px] px-1.5 rounded-full bg-[var(--surface2)] border border-[var(--border)] text-[10px] font-bold gap-1">
                      <span style="color: #3b82f6; font-size: 11px;">◈</span>
                      <span style="color: var(--muted);">${count}</span>
                    </div>
                  `;
              }

              const parent = c.getRow().getTreeParent() as { getData: () => CategoryRow } | false;
              if (parent) {
                const pData = parent.getData();
                if (pData.key === 'channel' || pData.key === 'product_class') {
                  return `<span style="font-size: 14px;">🔹</span>`;
                }
              }

              return c.getValue() ? `<span style="font-size: 14px;">📏</span>` : '';
            },
          },
          getCommonNameColumn('Name', (id, name, desc) => {
            setOverlayData({ id, name, desc });
          }),
          {
            title: 'Props',
            field: '_props',
            width: 76,
            hozAlign: 'center',
            headerHozAlign: 'center',
            formatter: (c: { getRow: () => { getData: () => CategoryRow } }) => {
              const d = c.getRow().getData();
              return `<button class="badge-btn badge-prop" data-rowid="${d.id}">⚙ ${(d._props || []).length}</button>`;
            },
          },
          {
            title: '',
            width: 64,
            hozAlign: 'center',
            headerSort: false,
            formatter: (c: { getRow: () => { getData: () => CategoryRow } }) => {
              const d = c.getRow().getData();
              const isCat = !!d._children;
              return `
                  <div class="flex items-center gap-1">
                    <button class="mc-edit" data-id="${d.id}" data-iscat="${isCat}" title="Edit">✎</button>
                    ${isCat ? `<button class="mc-add-el" data-catid="${d.id}" title="Add Element">＋</button>` : ''}
                  </div>
                `;
            },
          },
        ],
      });

      tabulatorInstance.current.on('dataTreeRowExpanded', (row: { getData: () => CategoryRow }) => {
        onToggleExpand(row.getData().id, true);
      });

      tabulatorInstance.current.on(
        'dataTreeRowCollapsed',
        (row: { getData: () => CategoryRow }) => {
          onToggleExpand(row.getData().id, false);
        }
      );

      tabulatorInstance.current.on('rowSelected', (row: { getData: () => CategoryRow }) => {
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
        const locBadge = target.closest('.badge-loc') as HTMLElement;
        const propBadge = target.closest('.badge-prop') as HTMLElement;

        if (editBtn) {
          const id = editBtn.dataset.id!;
          const isCat = editBtn.dataset.iscat === 'true';
          onEdit(id, isCat);
        } else if (locBadge) {
          const id = locBadge.dataset.rowid!;
          const row = tabulatorInstance.current?.getRow(id);
          const isCat = !!row?.getData()._children;
          onEdit(id, isCat, 'locales');
        } else if (propBadge) {
          const id = propBadge.dataset.rowid!;
          const row = tabulatorInstance.current?.getRow(id);
          const isCat = !!row?.getData()._children;
          onEdit(id, isCat, 'properties');
        } else if (addElBtn) {
          const catId = addElBtn.dataset.catid!;
          onAddElement(catId);
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
  }, [categories, onEdit, onAddElement, onToggleExpand, onSelect, expandedIds]);

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

  useEffect(() => {
    if (tabulatorInstance.current && targetElId && isTableReady) {
      const row = tabulatorInstance.current.getRow(targetElId);
      if (row) {
        const parent = row.getTreeParent();
        if (parent) parent.treeExpand();

        row.scrollTo();

        const el = typeof row.getElement === 'function' ? row.getElement() : null;
        if (el) {
          el.classList.add('highlight-row');
          setTimeout(() => el.classList.remove('highlight-row'), 2000);
        }

        onClearTarget();
      }
    }
  }, [targetElId, onClearTarget, isTableReady]);

  return (
    <div className="cat-view h-full flex flex-col animate-in fade-in duration-300 relative">
      <div ref={tableRef} id="cat-tabulator" className="flex-1 overflow-hidden"></div>

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
