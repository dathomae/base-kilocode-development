import React, { useState } from 'react';
import { BusinessOverview } from './BusinessOverview';
import { Plans } from './Plans';
import { PlanExecution } from './PlanExecution';
import { TreeNode } from '../../../types';
import { cn } from '../../../lib/utils';

interface PlanningViewProps {
  node: TreeNode;
  breadcrumbs: React.ReactNode;
  helpMsg: string;
  parentLabel: string;
  isHeaderMinimized: boolean;
  onToggleHeader: () => void;
}

const TABS = ['Execution', 'Plans', 'Trace'];

export const PlanningView: React.FC<PlanningViewProps> = ({
  node,
  breadcrumbs,
  helpMsg,
  parentLabel,
  isHeaderMinimized,
  onToggleHeader,
}) => {
  const [activeTab, setActiveTab] = useState(TABS[0]);

  const getBadgeStyles = (label: string) => {
    const l = label.toLowerCase();
    if (l.includes('production')) return 'bg-red-500/10 text-red-500 border-red-500/20';
    if (l.includes('staging')) return 'bg-yellow-500/10 text-yellow-500 border-yellow-500/20';
    return 'bg-[var(--surface2)] text-[var(--muted)] border-[var(--border)]';
  };

  const getCount = (tab: string) => {
    if (tab === 'Execution') return { count: 4, emoji: '📊' };
    if (tab === 'Plans') return { count: 7, emoji: '📋' };
    if (tab === 'Trace') return { count: 3, emoji: '🚀' };
    return { count: 0, emoji: '' };
  };

  return (
    <div
      className={cn(
        'detail-pane flex flex-col h-full bg-[var(--surface)] overflow-hidden transition-all duration-300',
        isHeaderMinimized ? 'p-6' : 'p-10'
      )}
    >
      <div
        className={cn(
          'breadcrumbs flex items-center justify-between shrink-0',
          isHeaderMinimized ? 'mb-4' : 'mb-6'
        )}
      >
        <div className="flex items-center gap-2 text-[11px] text-[var(--muted)] uppercase tracking-wider">
          {breadcrumbs}
        </div>
        <div className="flex items-center gap-3">
          <div
            className={cn(
              'text-[10px] py-0.5 px-2 rounded-full border font-bold uppercase tracking-widest leading-none',
              getBadgeStyles(parentLabel)
            )}
          >
            {parentLabel || node.type.replace('bucket_', '')}
          </div>
          <button
            onClick={onToggleHeader}
            className="p-1.5 rounded hover:bg-[var(--surface2)] text-[var(--muted)] hover:text-[var(--text)] transition-colors border border-[var(--border)] bg-[var(--surface)] cursor-pointer shadow-sm"
            title={isHeaderMinimized ? 'Expand Header' : 'Collapse Header'}
          >
            <svg
              width="14"
              height="14"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2.5"
              strokeLinecap="round"
              strokeLinejoin="round"
              className={cn(
                'transition-transform duration-300',
                isHeaderMinimized ? 'rotate-180' : ''
              )}
            >
              <polyline points="18 15 12 9 6 15"></polyline>
            </svg>
          </button>
        </div>
      </div>

      {!isHeaderMinimized && (
        <div className="detail-header flex flex-col gap-2.5 mb-10 shrink-0 animate-in fade-in slide-in-from-top-2 duration-300">
          <h1 className="text-[32px] font-semibold tracking-tight m-0">{node.label}</h1>
          <p className="text-[14px] text-[var(--muted)] max-w-[600px] leading-relaxed m-0">
            {helpMsg}
          </p>
        </div>
      )}

      <div
        className={cn(
          'flex items-center justify-between shrink-0',
          isHeaderMinimized ? 'mb-4' : 'mb-8'
        )}
      >
        <div className="bucket-tabs flex gap-2.5">
          {TABS.map(tab => {
            const { count, emoji } = getCount(tab);
            return (
              <button
                key={tab}
                className={cn(
                  'bucket-tab-btn py-1.5 px-5 rounded-full border border-[var(--border)] bg-none text-[13px] cursor-pointer text-[var(--muted)] transition-all hover:border-[var(--muted)] flex items-center gap-2',
                  activeTab === tab &&
                    'bg-[var(--text)] border-[var(--text)] text-[var(--surface)] font-medium'
                )}
                onClick={() => setActiveTab(tab)}
              >
                {tab}
                <span
                  className={cn(
                    'flex items-center justify-center min-w-[18px] h-[18px] px-1.5 rounded-full text-[10px] font-bold gap-1',
                    activeTab === tab
                      ? 'bg-[var(--muted)]/20 text-[var(--surface)]'
                      : 'bg-[var(--surface2)] text-[var(--muted)]'
                  )}
                >
                  <span>{emoji}</span>
                  {count}
                </span>
              </button>
            );
          })}
        </div>
      </div>

      <div className="bucket-tab-content flex-1 overflow-hidden">
        {activeTab === 'Execution' && <BusinessOverview />}
        {activeTab === 'Plans' && <Plans />}
        {activeTab === 'Trace' && <PlanExecution />}
      </div>
    </div>
  );
};
