import { X } from 'lucide-react';
import React, { useEffect, useRef, useState } from 'react';
import { TabulatorFull as Tabulator } from 'tabulator-tables';
import { attachMoreLinkListener } from '../../../lib/tabulatorUtils';
import { Converter } from '../../../types';

interface ConvertersProps {
  converters: Converter[];
  onEditConverter: (c: Converter) => void;
}

export const Converters: React.FC<ConvertersProps> = ({ converters, onEditConverter }) => {
  const tableRef = useRef<HTMLDivElement>(null);
  const tabulatorInstance = useRef<Tabulator | null>(null);
  const [isTableReady, setIsTableReady] = useState(false);
  const [overlayData, setOverlayData] = useState<{ id: string; name: string; desc: string } | null>(
    null
  );
  const [expandedRows, setExpandedRows] = useState<Set<string>>(new Set());

  useEffect(() => {
    if (tableRef.current && !tabulatorInstance.current) {
      setTimeout(() => setIsTableReady(false), 0);

      tabulatorInstance.current = new Tabulator(tableRef.current, {
        data: converters,
        layout: 'fitColumns',
        selectable: 1,
        placeholder: 'No converters defined',
        rowFormatter: (row: { getData: () => Converter; getElement: () => HTMLElement }) => {
          const data = row.getData();
          const isExpanded = expandedRows.has(data.id);

          // Remove existing detail element if any
          const existingDetail = row.getElement().querySelector('.row-detail');
          if (existingDetail) existingDetail.remove();

          if (isExpanded) {
            const detail = document.createElement('div');

            detail.classList.add(
              'row-detail',
              'p-4',
              'bg-[var(--surface2)]/30',
              'border-t',
              'border-[var(--border)]',
              'animate-in',
              'fade-in',
              'slide-in-from-top-1',
              'duration-200'
            );
            detail.style.width = '100%';

            // We need to render markdown here. Since Tabulator is DOM-based,
            // we'll use a temporary container and React to render into it if possible,
            // or just simple HTML. But since we want the full power of react-markdown,
            // we'll use a portal-like approach or just a simple innerHTML for now if it's too complex.
            // Actually, a better way is to use a custom component for the row detail.

            detail.innerHTML = `
              <div class="text-xs font-bold uppercase tracking-wider text-[var(--muted)] mb-2">Full Calculation</div>
              <div class="markdown-body p-3 bg-[var(--surface)] border border-[var(--border)] rounded-lg font-mono text-sm leading-relaxed">
                ${data.calculation.replace(/\n/g, '<br/>').replace(/`([^`]+)`/g, '<code class="px-1 py-0.5 bg-[var(--warn)]/10 text-[var(--warn)] rounded border border-[var(--warn)]/20">$1</code>')}
              </div>
            `;
            row.getElement().appendChild(detail);
          }
        },
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
            title: 'Type',
            field: 'type',
            width: 100,
            formatter: (c: { getValue: () => string }) => {
              const val = c.getValue();
              const color = val === 'Measure' ? 'var(--accent)' : 'var(--accent2)';
              return `<span class="px-2 py-0.5 rounded text-[10px] font-bold uppercase tracking-wider" style="background: ${color}20; color: ${color}; border: 1px solid ${color}40">${val}</span>`;
            },
          },
          {
            title: 'Assignment',
            field: 'assignment.name',
            widthGrow: 1,
            formatter: (c: { getRow: () => { getData: () => Converter } }) => {
              const data = c.getRow().getData().assignment;
              if (!data.name) return '';
              return `<span>${data.name} <span style="color: var(--muted); font-family: monospace;">( <span style="color: var(--warn); font-weight: bold;">${data.alias}</span> )</span></span>`;
            },
            sorter: 'string',
          },
          {
            title: 'Input',
            field: 'input.name',
            widthGrow: 1,
            formatter: (c: { getRow: () => { getData: () => Converter } }) => {
              const data = c.getRow().getData().input;
              if (!data.name) return '';
              return `<span>${data.name} <span style="color: var(--muted); font-family: monospace;">( <span style="color: var(--warn); font-weight: bold;">${data.alias}</span> )</span></span>`;
            },
            sorter: 'string',
          },
          {
            title: 'Calculation',
            field: 'calculation',
            widthGrow: 2,
            formatter: (c: { getRow: () => { getData: () => Converter } }) => {
              const data = c.getRow().getData();
              const val = data.calculation || '';
              const lines = val.split('\n');
              const firstLine = lines[0];
              const hasMore = lines.length > 1;
              const isExpanded = expandedRows.has(data.id);

              return `
                <div class="flex items-center justify-between gap-2 w-full">
                  <div class="flex items-center gap-2 font-mono text-[12px] text-[var(--muted)] truncate">
                    <span class="truncate">${firstLine}</span>
                  </div>
                  ${
                    hasMore
                      ? `
                    <button class="toggle-expand-btn flex items-center gap-1 px-2 py-0.5 bg-[var(--surface2)] border border-[var(--border)] rounded text-[10px] font-bold text-[var(--accent)] hover:bg-[var(--border)] transition-all" data-id="${data.id}">
                      ${isExpanded ? 'LESS' : 'MORE'}
                    </button>
                  `
                      : ''
                  }
                </div>
              `;
            },
          },
          {
            title: '',
            width: 40,
            headerSort: false,
            hozAlign: 'center',
            formatter: () =>
              `<button class="p-1 hover:bg-[var(--surface2)] rounded transition-colors text-[var(--muted)]"><svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path></svg></button>`,
            cellClick: (
              e: { stopPropagation: () => void },
              cell: { getRow: () => { getData: () => Converter } }
            ) => {
              e.stopPropagation();
              onEditConverter(cell.getRow().getData());
            },
          },
        ],
      });

      tabulatorInstance.current.on('tableBuilt', () => {
        setIsTableReady(true);
      });

      const tableEl = tableRef.current;
      const cleanup = attachMoreLinkListener(tableEl, (id, name, desc) => {
        setOverlayData({ id, name, desc });
      });

      const handleTableClick = (e: MouseEvent) => {
        const target = e.target as HTMLElement;
        const toggleBtn = target.closest('.toggle-expand-btn') as HTMLElement;
        if (toggleBtn) {
          const id = toggleBtn.dataset.id!;
          setExpandedRows(prev => {
            const next = new Set(prev);
            if (next.has(id)) next.delete(id);
            else next.add(id);
            return next;
          });
        }
      };
      tableEl?.addEventListener('click', handleTableClick);

      return () => {
        if (tabulatorInstance.current) {
          tabulatorInstance.current.destroy();
          tabulatorInstance.current = null;
        }
        cleanup();
        tableEl?.removeEventListener('click', handleTableClick);
      };
    }
  }, [converters, onEditConverter, expandedRows]);

  useEffect(() => {
    if (tabulatorInstance.current && isTableReady) {
      tabulatorInstance.current.setData(converters);
    }
  }, [converters, isTableReady]);

  return (
    <div className="flex flex-col h-full overflow-hidden animate-in fade-in duration-500">
      <div className="flex-1 overflow-hidden border border-[var(--border)] rounded-xl bg-[var(--surface)] shadow-sm relative">
        <div ref={tableRef} className="h-full" />
        {!isTableReady && (
          <div className="absolute inset-0 flex items-center justify-center bg-[var(--surface)]/50 backdrop-blur-[1px] z-10">
            <div className="w-6 h-6 border-2 border-[var(--accent)] border-t-transparent rounded-full animate-spin" />
          </div>
        )}
      </div>

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
