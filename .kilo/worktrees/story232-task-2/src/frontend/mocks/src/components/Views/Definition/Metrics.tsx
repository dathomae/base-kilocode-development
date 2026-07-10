import React, { useEffect, useRef, useState } from 'react';
import { TabulatorFull as Tabulator } from 'tabulator-tables';
import { Metric } from '../../../types';

interface MetricsProps {
  metrics: Metric[];
  onAdd: () => void;
  onEdit: (m: Metric) => void;
  onReorder: (id: number, direction: 'up' | 'down') => void;
  selectedId: string | null;
  onSelect: (id: string | null) => void;
}

export const Metrics: React.FC<MetricsProps> = ({
  metrics,
  onAdd,
  onEdit,
  onReorder,
  selectedId,
  onSelect,
}) => {
  const tableRef = useRef<HTMLDivElement>(null);
  const tabulatorInstance = useRef<Tabulator | null>(null);
  const [isTableReady, setIsTableReady] = useState(false);

  useEffect(() => {
    if (tableRef.current && !tabulatorInstance.current) {
      // eslint-disable-next-line react-hooks/set-state-in-effect
      setIsTableReady(false);

      tabulatorInstance.current = new Tabulator(tableRef.current, {
        data: metrics,
        layout: 'fitColumns',
        selectable: 1,
        placeholder: 'No metrics configured',
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
            title: 'Measure',
            field: 'measure',
            widthGrow: 1,
            formatter: (c: { getValue: () => string }) =>
              `<span class="font-medium text-[var(--text)]">${c.getValue()}</span>`,
          },
          {
            title: 'Dimensions',
            field: 'dims',
            widthGrow: 1.5,
            formatter: (c: { getValue: () => string[] }) => {
              const dims = c.getValue() || [];
              return `<div class="flex flex-wrap gap-1">${dims.map((d: string) => `<span class="px-1.5 py-0.5 rounded bg-[var(--surface2)] border border-[var(--border)] text-[10px] text-[var(--muted)]">${d}</span>`).join('')}</div>`;
            },
          },
          {
            title: 'Grain',
            field: 'grain',
            width: 100,
            formatter: (c: { getValue: () => string }) =>
              `<span class="capitalize text-[var(--muted)]">${c.getValue()}</span>`,
          },
          {
            title: '',
            width: 100,
            headerSort: false,
            hozAlign: 'center',
            formatter: () => `
              <div class="flex items-center gap-1">
                <button class="p-1 hover:bg-[var(--surface2)] rounded transition-colors text-[var(--muted)] mc-up"><svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="18 15 12 9 6 15"></polyline></svg></button>
                <button class="p-1 hover:bg-[var(--surface2)] rounded transition-colors text-[var(--muted)] mc-down"><svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="6 9 12 15 18 9"></polyline></svg></button>
                <button class="p-1 hover:bg-[var(--surface2)] rounded transition-colors text-[var(--muted)] mc-edit"><svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path></svg></button>
              </div>
            `,
            cellClick: (
              e: { stopPropagation: () => void; target: unknown },
              cell: { getRow: () => { getData: () => Metric } }
            ) => {
              e.stopPropagation();
              const target = e.target as HTMLElement;
              const upBtn = target.closest('.mc-up');
              const downBtn = target.closest('.mc-down');
              const editBtn = target.closest('.mc-edit');
              const data = cell.getRow().getData();

              if (upBtn) onReorder(data.id, 'up');
              else if (downBtn) onReorder(data.id, 'down');
              else if (editBtn) onEdit(data);
            },
          },
        ],
      });

      tabulatorInstance.current.on('tableBuilt', () => {
        setIsTableReady(true);
      });

      tabulatorInstance.current.on('rowSelected', (row: { getData: () => Metric }) => {
        onSelect(row.getData().id.toString());
      });

      tabulatorInstance.current.on('rowDeselected', () => {
        if (tabulatorInstance.current?.getSelectedRows().length === 0) {
          onSelect(null);
        }
      });

      return () => {
        if (tabulatorInstance.current) {
          tabulatorInstance.current.destroy();
          tabulatorInstance.current = null;
        }
      };
    }
  }, [metrics, onEdit, onReorder, onSelect, isTableReady]);

  useEffect(() => {
    if (tabulatorInstance.current && isTableReady) {
      tabulatorInstance.current.setData(metrics);
    }
  }, [metrics, isTableReady]);

  useEffect(() => {
    if (tabulatorInstance.current && isTableReady && selectedId) {
      tabulatorInstance.current.deselectRow();
      tabulatorInstance.current.selectRow(selectedId);
    }
  }, [selectedId, isTableReady]);

  return (
    <div className="metrics-view flex flex-col h-full animate-in fade-in duration-300">
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-[14px] font-semibold m-0 text-[var(--muted)] uppercase tracking-wider">
          Metrics Configuration
        </h3>
        <button
          onClick={onAdd}
          className="py-1 px-3 rounded-md bg-[var(--accent2)] text-[#0f1117] text-[12px] font-bold border-none cursor-pointer hover:opacity-90 transition-opacity"
        >
          + Add Metric
        </button>
      </div>

      <div className="flex-1 overflow-hidden border border-[var(--border)] rounded-lg bg-[var(--surface)] relative">
        <div ref={tableRef} className="h-full" />
        {!isTableReady && (
          <div className="absolute inset-0 flex items-center justify-center bg-[var(--surface)]/50 backdrop-blur-[1px] z-10">
            <div className="w-6 h-6 border-2 border-[var(--accent)] border-t-transparent rounded-full animate-spin" />
          </div>
        )}
      </div>
    </div>
  );
};
