import React, { useState } from 'react';
import { PlanTypes } from './PlanTypes';
import { Categories } from './Categories';
import { Taxonomies } from './Taxonomies';
import { Plan, Metric, Category, Taxonomy, TreeNode } from '../../../types';
import { cn } from '../../../lib/utils';

interface DefinitionViewProps {
  node: TreeNode;
  breadcrumbs: React.ReactNode;
  helpMsg: string;
  parentLabel: string;
  isHeaderMinimized: boolean;
  onToggleHeader: () => void;
  // Plan Types props
  plans: Plan[];
  onAddPlan: () => void;
  onEditPlan: (p: Plan) => void;
  onDeletePlan: (id: number) => void;
  getMetricsForPlan: (id: number, label: string) => Metric[];
  onAddMetric: () => void;
  onEditMetric: (m: Metric) => void;
  onReorderMetric: (id: number, direction: 'up' | 'down') => void;
  selectedMetricId: string | null;
  onSelectMetric: (id: string | null) => void;
  activePlanId: number | null;
  setActivePlanId: (id: number | null) => void;
  // Categories props
  categories: Category[];
  expandedCatIds: string[];
  onToggleCatExpand: (id: string, expanded: boolean) => void;
  selectedCatId: string | null;
  onSelectCat: (id: string | null) => void;
  targetElId: string | null;
  onClearTarget: () => void;
  onAddCategory: () => void;
  onAddElementToCat: (catId: string) => void;
  onEditCatOrEl: (id: string, isCat: boolean, tab?: 'general' | 'properties' | 'locales') => void;
  // Taxonomies props
  taxonomies: Taxonomy[];
  expandedTaxIds: string[];
  onToggleTaxExpand: (id: string, expanded: boolean) => void;
  selectedTaxId: string | null;
  onSelectTax: (id: string | null) => void;
  onAddTaxonomy: () => void;
  onEditTaxonomy: (id: string, tab?: 'general' | 'properties' | 'locales' | 'dag') => void;
  onAddTaxElement: (taxId: string, parentNodeId?: string) => void;
  onGoToCategory: (catId: string, elId?: string) => void;
}

const TABS = ['Plan Types', 'Taxonomies', 'Categories'];

export const DefinitionView: React.FC<DefinitionViewProps> = ({
  node,
  breadcrumbs,
  helpMsg,
  parentLabel,
  isHeaderMinimized,
  onToggleHeader,
  plans,
  onAddPlan,
  onEditPlan,
  onDeletePlan,
  getMetricsForPlan,
  onAddMetric,
  onEditMetric,
  onReorderMetric,
  selectedMetricId,
  onSelectMetric,
  activePlanId,
  setActivePlanId,
  categories,
  expandedCatIds,
  onToggleCatExpand,
  selectedCatId,
  onSelectCat,
  targetElId,
  onClearTarget,
  onAddCategory,
  onAddElementToCat,
  onEditCatOrEl,
  taxonomies,
  expandedTaxIds,
  onToggleTaxExpand,
  selectedTaxId,
  onSelectTax,
  onAddTaxonomy,
  onEditTaxonomy,
  onAddTaxElement,
  onGoToCategory,
}) => {
  const [activeTab, setActiveTab] = useState(TABS[0]);

  const getBadgeStyles = (label: string) => {
    const l = label.toLowerCase();
    if (l.includes('production')) return 'bg-red-500/10 text-red-500 border-red-500/20';
    if (l.includes('staging')) return 'bg-yellow-500/10 text-yellow-500 border-yellow-500/20';
    return 'bg-[var(--surface2)] text-[var(--muted)] border-[var(--border)]';
  };

  const renderActiveActionButton = () => {
    if (activeTab === 'Plan Types' && activePlanId === null) {
      return (
        <button
          onClick={onAddPlan}
          className="py-1.5 px-4 rounded-full bg-[var(--accent)] text-white text-[13px] font-medium border-none cursor-pointer hover:opacity-90 transition-opacity"
        >
          + Add Plan Type
        </button>
      );
    }
    if (activeTab === 'Categories') {
      return (
        <button
          className="py-1.5 px-4 rounded-full bg-[var(--accent)] text-white text-[13px] font-medium border-none cursor-pointer hover:opacity-90 transition-opacity"
          onClick={onAddCategory}
        >
          ＋ Add Category
        </button>
      );
    }
    if (activeTab === 'Taxonomies') {
      return (
        <button
          className="py-1.5 px-4 rounded-full bg-[var(--accent)] text-white text-[13px] font-medium border-none cursor-pointer hover:opacity-90 transition-opacity"
          onClick={onAddTaxonomy}
        >
          ＋ Add Taxonomy
        </button>
      );
    }
    return null;
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
            const isActive = activeTab === tab;
            return (
              <button
                key={tab}
                className={cn(
                  'bucket-tab-btn py-1.5 px-5 rounded-full border border-[var(--border)] bg-none text-[13px] cursor-pointer text-[var(--muted)] transition-all hover:border-[var(--muted)] flex items-center gap-2',
                  isActive &&
                    'bg-[var(--text)] border-[var(--text)] text-[var(--surface)] font-medium'
                )}
                onClick={() => setActiveTab(tab)}
              >
                {tab}
                <div className="flex items-center gap-1.5 ml-1">
                  {tab === 'Plan Types' && (
                    <span
                      className={cn(
                        'flex items-center justify-center min-w-[18px] h-[18px] px-1.5 rounded-full text-[10px] font-bold gap-1',
                        isActive
                          ? 'bg-[var(--muted)]/20 text-[var(--surface)]'
                          : 'bg-[var(--surface2)] text-[var(--muted)]'
                      )}
                    >
                      <span>📋</span>
                      {plans.length}
                    </span>
                  )}
                  {tab === 'Taxonomies' && (
                    <>
                      <span
                        className={cn(
                          'flex items-center justify-center min-w-[18px] h-[18px] px-1.5 rounded-full text-[10px] font-bold gap-1',
                          isActive
                            ? 'bg-[var(--muted)]/20 text-[var(--surface)]'
                            : 'bg-[var(--surface2)] text-[var(--muted)]'
                        )}
                        title="Taxonomies"
                      >
                        <span>🏷️</span>
                        {taxonomies.length}
                      </span>
                      <span
                        className={cn(
                          'flex items-center justify-center min-w-[18px] h-[18px] px-1.5 rounded-full text-[10px] font-bold gap-1',
                          isActive
                            ? 'bg-[var(--muted)]/20 text-[var(--surface)]'
                            : 'bg-[var(--surface2)] text-[var(--muted)]'
                        )}
                        title="Dimensions"
                      >
                        <span className="text-[#3b82f6] text-[12px]">⎇</span>
                        {taxonomies.filter(t => t.isMeasuresDimension).length}
                      </span>
                    </>
                  )}
                  {tab === 'Categories' && (
                    <>
                      <span
                        className={cn(
                          'flex items-center justify-center min-w-[18px] h-[18px] px-1.5 rounded-full text-[10px] font-bold gap-1',
                          isActive
                            ? 'bg-[var(--muted)]/20 text-[var(--surface)]'
                            : 'bg-[var(--surface2)] text-[var(--muted)]'
                        )}
                        title="Categories"
                      >
                        <span>📁</span>
                        {categories.length}
                      </span>
                      <span
                        className={cn(
                          'flex items-center justify-center min-w-[18px] h-[18px] px-1.5 rounded-full text-[10px] font-bold gap-1',
                          isActive
                            ? 'bg-[var(--muted)]/20 text-[var(--surface)]'
                            : 'bg-[var(--surface2)] text-[var(--muted)]'
                        )}
                        title="Units"
                      >
                        <span className="text-[#3b82f6] text-[12px]">◈</span>
                        {categories.reduce((acc, cat) => {
                          if (cat.key === 'channel' || cat.key === 'product_class') {
                            return acc + 1;
                          }
                          return acc + (cat.isBaseUnit ? 1 : 0);
                        }, 0)}
                      </span>
                    </>
                  )}
                </div>
              </button>
            );
          })}
        </div>
        <div className="bucket-actions">{renderActiveActionButton()}</div>
      </div>

      <div className="bucket-tab-content flex-1 overflow-hidden">
        {activeTab === 'Plan Types' && (
          <PlanTypes
            plans={plans}
            onAdd={onAddPlan}
            onEdit={onEditPlan}
            onDelete={onDeletePlan}
            getMetricsForPlan={getMetricsForPlan}
            onAddMetric={onAddMetric}
            onEditMetric={onEditMetric}
            onReorderMetric={onReorderMetric}
            selectedMetricId={selectedMetricId}
            onSelectMetric={onSelectMetric}
            activePlanId={activePlanId}
            setActivePlanId={setActivePlanId}
            isHeaderMinimized={isHeaderMinimized}
          />
        )}
        {activeTab === 'Categories' && (
          <Categories
            categories={categories}
            expandedIds={expandedCatIds}
            onToggleExpand={onToggleCatExpand}
            selectedId={selectedCatId}
            onSelect={onSelectCat}
            targetElId={targetElId}
            onClearTarget={onClearTarget}
            onAdd={onAddCategory}
            onAddElement={onAddElementToCat}
            onEdit={onEditCatOrEl}
          />
        )}
        {activeTab === 'Taxonomies' && (
          <Taxonomies
            taxonomies={taxonomies}
            categories={categories}
            expandedIds={expandedTaxIds}
            onToggleExpand={onToggleTaxExpand}
            selectedId={selectedTaxId}
            onSelect={onSelectTax}
            onAdd={onAddTaxonomy}
            onEdit={onEditTaxonomy}
            onAddElement={onAddTaxElement}
            onGoToCategory={onGoToCategory}
          />
        )}
      </div>
    </div>
  );
};
