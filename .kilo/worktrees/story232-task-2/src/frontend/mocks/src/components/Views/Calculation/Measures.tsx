import { X } from 'lucide-react';
import React, { useEffect, useRef, useState } from 'react';
import { TabulatorFull as Tabulator } from 'tabulator-tables';
import { attachMoreLinkListener, getCommonNameColumn } from '../../../lib/tabulatorUtils';
import { Category, Measure, Taxonomy } from '../../../types';

interface MeasuresProps {
  categories: Category[];
  taxonomies: Taxonomy[];
  measures: Measure[];
  onMeasuresChange: (measures: Measure[]) => void;
  onEditMeasure: (m: Measure, tab?: 'general' | 'locales' | 'properties') => void;
}

export const Measures: React.FC<MeasuresProps> = props => {
  const { measures, onEditMeasure } = props;

  const tableRef = useRef<HTMLDivElement>(null);
  const tabulatorInstance = useRef<Tabulator | null>(null);
  const [isTableReady, setIsTableReady] = useState(false);

  // Overlay state
  const [overlayData, setOverlayData] = useState<{ id: string; name: string; desc: string } | null>(
    null
  );

  useEffect(() => {
    if (tableRef.current && !tabulatorInstance.current) {
      setTimeout(() => setIsTableReady(false), 0);

      tabulatorInstance.current = new Tabulator(tableRef.current, {
        data: measures,
        dataTree: true,
        dataTreeChildField: '_children',
        dataTreeChildIndent: 24,
        dataTreeElementColumn: 'name',
        layout: 'fitColumns',
        selectable: 1,
        placeholder: 'No measures defined',
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
            ...getCommonNameColumn('Measure', (id, name, desc) => {
              setOverlayData({ id, name, desc });
            }),
            widthGrow: 1,
          },
          {
            title: 'Data Type',
            field: 'dataType',
            width: 120,
            formatter: (c: { getValue: () => string }) =>
              `<span class="px-2 py-0.5 bg-[rgba(245,166,35,0.08)] border border-[rgba(245,166,35,0.25)] rounded text-[10px] font-mono text-[var(--warn)]">${c.getValue()}</span>`,
          },
          {
            title: 'Units',
            field: 'units',
            widthGrow: 4,
            formatter: (c: { getValue: () => string }) => {
              const val = c.getValue() || '';
              const parts = val.split(/(\s+|\*|\/|\^)/);
              const styled = parts
                .map((part: string) => {
                  if (['*', '/', '^'].includes(part.trim())) {
                    return `<span style="color: var(--warn)">${part}</span>`;
                  }
                  if (part.trim() && !['*', '/', '^'].includes(part.trim())) {
                    return `<span style="color: var(--accent2)">${part}</span>`;
                  }
                  return part;
                })
                .join('');
              return `<span class="font-mono text-[12px]">${styled}</span>`;
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
              cell: { getRow: () => { getData: () => Measure } }
            ) => {
              e.stopPropagation();
              onEditMeasure(cell.getRow().getData());
            },
          },
        ],
      });

      tabulatorInstance.current.on('tableBuilt', () => {
        setIsTableReady(true);
      });

      const tableEl = tableRef.current;
      const handleClick = (e: MouseEvent) => {
        const target = e.target as HTMLElement;
        const locBadge = target.closest('.badge-loc') as HTMLElement;
        if (locBadge) {
          const id = locBadge.dataset.id!;
          const measure = measures.find(m => m.id === id);
          if (measure) onEditMeasure(measure, 'locales');
        }
      };

      tableEl.addEventListener('click', handleClick);

      // Attach the "More..." link listener
      const cleanup = attachMoreLinkListener(tableEl, (id, name, desc) => {
        setOverlayData({ id, name, desc });
      });

      return () => {
        tableEl.removeEventListener('click', handleClick);
        if (tabulatorInstance.current) {
          tabulatorInstance.current.destroy();
          tabulatorInstance.current = null;
        }
        cleanup();
      };
    }
  }, [measures, onEditMeasure]); // Initialize only once

  useEffect(() => {
    if (tabulatorInstance.current && isTableReady) {
      tabulatorInstance.current.setData(measures);
    }
  }, [measures, isTableReady]);

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
