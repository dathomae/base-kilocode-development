import { X } from 'lucide-react';
import React, { useEffect, useRef, useState } from 'react';
import { TabulatorFull as Tabulator } from 'tabulator-tables';
import { attachMoreLinkListener, getCommonNameColumn } from '../../../lib/tabulatorUtils';
import { cn } from '../../../lib/utils';
import { Metric, Plan } from '../../../types';
import { FlowType, FlowView } from './FlowView';
import { Metrics } from './Metrics';

interface PlanTypesProps {
  plans: Plan[];
  onAdd: () => void;
  onEdit: (plan: Plan) => void;
  onDelete: (id: number) => void;
  getMetricsForPlan: (id: number, label: string) => Metric[];
  onAddMetric: () => void;
  onEditMetric: (m: Metric) => void;
  onReorderMetric: (id: number, direction: 'up' | 'down') => void;
  selectedMetricId: string | null;
  onSelectMetric: (id: string | null) => void;
  activePlanId: number | null;
  setActivePlanId: (id: number | null) => void;
  isHeaderMinimized?: boolean;
}

export const PlanTypes: React.FC<PlanTypesProps> = ({
  plans,
  onEdit,
  onDelete,
  getMetricsForPlan,
  onAddMetric,
  onEditMetric,
  onReorderMetric,
  selectedMetricId,
  onSelectMetric,
  activePlanId,
  setActivePlanId,
  isHeaderMinimized = false,
}) => {
  const tableRef = useRef<HTMLDivElement>(null);
  const tabulatorInstance = useRef<Tabulator | null>(null);
  const [isTableReady, setIsTableReady] = useState(false);
  const [activeTab, setActiveTab] = useState<'list' | 'flow'>('list');
  const [flowType, setFlowType] = useState<FlowType>('Sankey');
  const [searchQuery, setSearchQuery] = useState('');
  const [overlayData, setOverlayData] = useState<{ id: string; name: string; desc: string } | null>(
    null
  );

  useEffect(() => {
    if (tableRef.current && !activePlanId) {
      setTimeout(() => setIsTableReady(false), 0);

      tabulatorInstance.current = new Tabulator(tableRef.current, {
        data: plans,
        dataTree: true,
        dataTreeChildField: '_children',
        dataTreeChildIndent: 24,
        dataTreeElementColumn: 'label',
        layout: 'fitColumns',
        selectable: 1,
        placeholder: 'No plan types defined',
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
          getCommonNameColumn(
            'Plan Type',
            (id, name, desc) => {
              setOverlayData({ id, name, desc });
            },
            'label'
          ),
          {
            title: 'Kind',
            field: 'kind',
            width: 100,
            formatter: (c: { getValue: () => string }) => {
              const val = c.getValue();
              const styles =
                val === 'forecast'
                  ? 'bg-[var(--accent2)]/10 text-[var(--accent2)] border-[var(--accent2)]/20'
                  : 'bg-[var(--accent)]/10 text-[var(--accent)] border-[var(--accent)]/20';
              return `<span class="px-2 py-0.5 rounded-full text-[10px] font-bold uppercase tracking-wider border ${styles}">${val}</span>`;
            },
          },
          {
            title: 'Metrics',
            field: 'id',
            width: 80,
            hozAlign: 'center',
            formatter: (c: { getRow: () => { getData: () => Plan } }) => {
              const plan = c.getRow().getData();
              const metricsCount = getMetricsForPlan(plan.id, plan.label).length;
              return `<button class="badge-btn badge-prop">⚙ ${metricsCount}</button>`;
            },
          },
        ],
      });

      tabulatorInstance.current.on('tableBuilt', () => {
        setIsTableReady(true);
      });

      tabulatorInstance.current.on(
        'rowClick',
        (e: { target: unknown }, row: { getData: () => Plan }) => {
          // Only trigger if not clicking a button or link
          if (
            !(e.target as HTMLElement).closest('button') &&
            !(e.target as HTMLElement).closest('a')
          ) {
            setActivePlanId(row.getData().id);
          }
        }
      );

      const tableEl = tableRef.current;
      const handleClick = (e: MouseEvent) => {
        const target = e.target as HTMLElement;
        const locBadge = target.closest('.badge-loc') as HTMLElement;
        if (locBadge) {
          const id = parseInt(locBadge.dataset.id!);
          const plan = plans.find(p => p.id === id);
          if (plan) onEdit(plan);
        }
      };

      tableEl.addEventListener('click', handleClick);

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
  }, [activePlanId, plans, onEdit, getMetricsForPlan, setActivePlanId]);

  useEffect(() => {
    if (tabulatorInstance.current && isTableReady && !activePlanId) {
      tabulatorInstance.current.setData(plans);
    }
  }, [plans, isTableReady, activePlanId]);

  useEffect(() => {
    if (tabulatorInstance.current && isTableReady && !activePlanId) {
      tabulatorInstance.current.setFilter('label', 'like', searchQuery);
    }
  }, [searchQuery, isTableReady, activePlanId]);

  const selectedPlan = plans.find(p => p.id === activePlanId);

  if (selectedPlan && activePlanId !== null) {
    return (
      <div className="plan-detail-view flex flex-col h-full animate-in fade-in slide-in-from-right-4 duration-300 overflow-hidden">
        <div
          className={cn(
            'flex items-center gap-6 shrink-0 animate-in fade-in slide-in-from-top-2 duration-300',
            isHeaderMinimized ? 'mb-4' : 'mb-8'
          )}
        >
          <button
            onClick={() => setActivePlanId(null)}
            className={cn(
              'flex items-center justify-center rounded-full hover:bg-[var(--surface2)] border border-[var(--border)] bg-[var(--surface)] cursor-pointer text-[var(--muted)] hover:text-[var(--text)] transition-all shadow-sm',
              isHeaderMinimized ? 'w-8 h-8' : 'w-10 h-10'
            )}
          >
            <svg
              width={isHeaderMinimized ? '16' : '20'}
              height={isHeaderMinimized ? '16' : '20'}
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2.5"
              strokeLinecap="round"
              strokeLinejoin="round"
            >
              <line x1="19" y1="12" x2="5" y2="12"></line>
              <polyline points="12 19 5 12 12 5"></polyline>
            </svg>
          </button>
          <div className="flex flex-col gap-0.5">
            <div className="flex items-center gap-3">
              <h1
                className={cn(
                  'font-semibold tracking-tight m-0',
                  isHeaderMinimized ? 'text-[24px]' : 'text-[32px]'
                )}
              >
                {selectedPlan.label}
              </h1>
              <span
                className={cn(
                  'px-2.5 py-0.5 rounded-full text-[11px] font-bold uppercase tracking-wider border',
                  selectedPlan.kind === 'forecast'
                    ? 'bg-[var(--accent2)]/10 text-[var(--accent2)] border-[var(--accent2)]/20'
                    : 'bg-[var(--accent)]/10 text-[var(--accent)] border-[var(--accent)]/20'
                )}
              >
                {selectedPlan.kind}
              </span>
            </div>
            {!isHeaderMinimized && (
              <div className="flex items-center gap-2 text-[13px] text-[var(--muted)]">
                <span className="font-mono font-bold text-[var(--accent2)]">
                  {selectedPlan.key}
                </span>
                <span>•</span>
                <span>{selectedPlan.desc || 'No description provided.'}</span>
              </div>
            )}
            {isHeaderMinimized && (
              <div className="text-[11px] text-[var(--muted)] font-mono font-bold text-[var(--accent2)]">
                {selectedPlan.key}
              </div>
            )}
          </div>
          <div className="ml-auto flex gap-3">
            <button
              onClick={() => {
                if (confirm('Are you sure you want to delete this plan type?')) {
                  onDelete(selectedPlan.id);
                  setActivePlanId(null);
                }
              }}
              className={cn(
                'rounded-full border border-red-500/20 bg-red-500/5 text-red-500 font-medium cursor-pointer hover:bg-red-500/10 transition-all shadow-sm',
                isHeaderMinimized ? 'py-1.5 px-4 text-[12px]' : 'py-2 px-5 text-[13px]'
              )}
            >
              Delete
            </button>
          </div>
        </div>

        <div
          className={cn(
            'tabs flex items-center justify-between border-b border-[var(--border)] shrink-0',
            isHeaderMinimized ? 'mb-4' : 'mb-8'
          )}
        >
          <div className="flex gap-2">
            <button
              className={cn(
                'tab-btn p-2 px-4.5 bg-none border-none cursor-pointer text-[13px] text-[var(--muted)] border-b-2 border-transparent -mb-px transition-all hover:text-[var(--text)] flex items-center gap-2',
                activeTab === 'list' &&
                  'text-[var(--accent2)] border-b-[var(--accent2)] font-semibold'
              )}
              onClick={() => setActiveTab('list')}
            >
              List
              <span
                className={cn(
                  'flex items-center justify-center min-w-[18px] h-[18px] px-1.5 rounded-full text-[10px] font-bold gap-1',
                  activeTab === 'list'
                    ? 'bg-[var(--muted)]/20 text-[var(--surface)]'
                    : 'bg-[var(--surface2)] text-[var(--muted)]'
                )}
              >
                <span>📏</span>
                {getMetricsForPlan(selectedPlan.id, selectedPlan.label).length}
              </span>
            </button>
            <button
              className={cn(
                'tab-btn p-2 px-4.5 bg-none border-none cursor-pointer text-[13px] text-[var(--muted)] border-b-2 border-transparent -mb-px transition-all hover:text-[var(--text)] flex items-center gap-2',
                activeTab === 'flow' &&
                  'text-[var(--accent2)] border-b-[var(--accent2)] font-semibold'
              )}
              onClick={() => setActiveTab('flow')}
            >
              Flow
            </button>
          </div>

          {activeTab === 'flow' && (
            <div className="flex items-center gap-1 bg-[var(--surface2)] p-1 rounded-lg border border-[var(--border)] mb-1 animate-in fade-in zoom-in-95 duration-300">
              {(['Sankey', 'Tree', 'Radial', 'Sequence'] as FlowType[]).map(type => (
                <button
                  key={type}
                  onClick={() => setFlowType(type)}
                  className={cn(
                    'px-3 py-1 rounded-md text-[11px] font-medium transition-all cursor-pointer',
                    flowType === type
                      ? 'bg-[var(--accent2)] text-white shadow-sm'
                      : 'text-[var(--muted)] hover:text-[var(--text)] hover:bg-[var(--surface)]'
                  )}
                >
                  {type}
                </button>
              ))}
            </div>
          )}
        </div>

        <div className="flex-1 overflow-y-auto min-h-0">
          {activeTab === 'list' ? (
            <Metrics
              metrics={getMetricsForPlan(selectedPlan.id, selectedPlan.label)}
              onAdd={onAddMetric}
              onEdit={onEditMetric}
              onReorder={onReorderMetric}
              selectedId={selectedMetricId}
              onSelect={onSelectMetric}
            />
          ) : (
            <FlowView type={flowType} />
          )}
        </div>
      </div>
    );
  }

  return (
    <div className="plan-types-view flex flex-col h-full animate-in fade-in duration-300 relative">
      <div className="flex items-center justify-between mb-4 shrink-0">
        <div className="relative w-64">
          <svg
            className="absolute left-3 top-1/2 -translate-y-1/2 text-[var(--muted)]"
            width="14"
            height="14"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <circle cx="11" cy="11" r="8"></circle>
            <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
          </svg>
          <input
            type="text"
            placeholder="Search plan types..."
            value={searchQuery}
            onChange={e => setSearchQuery(e.target.value)}
            className="w-full bg-[var(--surface)] border border-[var(--border)] rounded-lg pl-9 pr-4 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-[var(--accent)]/50 transition-all"
          />
        </div>
      </div>

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
