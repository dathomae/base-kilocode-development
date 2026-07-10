import React, { useState } from 'react';
import { Measures } from './Measures';
import { Spreaders } from './Spreaders';

import { Plus } from 'lucide-react';
import { cn } from '../../../lib/utils';
import { Category, Converter, Measure, Taxonomy, TreeNode } from '../../../types';
import { ConverterModal } from '../../Modals/ConverterModal';
import { MeasureModal } from '../../Modals/MeasureModal';
import { Aggregators } from './Aggregators';
import { Converters } from './Converters';

interface CalculationViewProps {
  node: TreeNode;
  breadcrumbs: React.ReactNode;
  helpMsg: string;
  parentLabel: string;
  isHeaderMinimized: boolean;
  onToggleHeader: () => void;
  categories: Category[];
  taxonomies: Taxonomy[];
  measures: Measure[];
  onMeasuresChange: (measures: Measure[]) => void;
  converters: Converter[];
  onConvertersChange: (converters: Converter[]) => void;
}

const TABS = ['Measures', 'Spreaders', 'Aggregators', 'Converters'];

export const CalculationView: React.FC<CalculationViewProps> = ({
  node,
  breadcrumbs,
  helpMsg,
  parentLabel,
  isHeaderMinimized,
  onToggleHeader,
  categories,
  taxonomies,
  measures,
  onMeasuresChange,
  converters,
  onConvertersChange,
}) => {
  const [activeTab, setActiveTab] = useState(TABS[0]);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingMeasure, setEditingMeasure] = useState<Measure | null>(null);
  const [initialMeasureTab, setInitialMeasureTab] = useState<'general' | 'locales' | 'properties'>(
    'general'
  );

  const [isConverterModalOpen, setIsConverterModalOpen] = useState(false);
  const [editingConverter, setEditingConverter] = useState<Converter | null>(null);

  const getBadgeStyles = (label: string) => {
    const l = label.toLowerCase();
    if (l.includes('production')) return 'bg-red-500/10 text-red-500 border-red-500/20';
    if (l.includes('staging')) return 'bg-yellow-500/10 text-yellow-500 border-yellow-500/20';
    return 'bg-[var(--surface2)] text-[var(--muted)] border-[var(--border)]';
  };

  const getCount = (tab: string) => {
    if (tab === 'Measures') return { count: measures.length, emoji: '📏' };
    if (tab === 'Spreaders') return { count: 4, emoji: '👐' };
    if (tab === 'Aggregators') return { count: 6, emoji: '🤏' };
    if (tab === 'Converters') return { count: converters.length, emoji: '🔄' };
    return { count: 0, emoji: '' };
  };

  const handleMeasureConfirm = (data: Measure) => {
    if (editingMeasure) {
      onMeasuresChange(measures.map(m => (m.id === data.id ? data : m)));
    } else {
      onMeasuresChange([...measures, data]);
    }
    setIsModalOpen(false);
  };

  const handleMeasureDelete = (id: string) => {
    onMeasuresChange(measures.filter(m => m.id !== id));
    setIsModalOpen(false);
  };

  const handleConverterConfirm = (data: Converter) => {
    if (editingConverter) {
      onConvertersChange(converters.map(c => (c.id === data.id ? data : c)));
    } else {
      onConvertersChange([...converters, data]);
    }
    setIsConverterModalOpen(false);
  };

  const handleConverterDelete = (id: string) => {
    onConvertersChange(converters.filter(c => c.id !== id));
    setIsConverterModalOpen(false);
  };

  const renderActiveActionButton = () => {
    if (activeTab === 'Measures') {
      return (
        <button
          onClick={() => {
            setEditingMeasure(null);
            setIsModalOpen(true);
          }}
          className="flex items-center gap-2 px-5 py-1.5 bg-[#3b82f6] text-white rounded-full text-[13px] font-medium hover:opacity-90 transition-all shadow-sm border-none cursor-pointer"
        >
          <Plus size={14} />
          Add Measure
        </button>
      );
    }
    if (activeTab === 'Converters') {
      return (
        <button
          onClick={() => {
            setEditingConverter(null);
            setIsConverterModalOpen(true);
          }}
          className="flex items-center gap-2 px-5 py-1.5 bg-[#3b82f6] text-white rounded-full text-[13px] font-medium hover:opacity-90 transition-all shadow-sm border-none cursor-pointer"
        >
          <Plus size={14} />
          Add Converter
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
        <div className="bucket-tabs flex items-center gap-2.5">
          {TABS.map(tab => {
            const { count, emoji } = getCount(tab);
            const isActive = activeTab === tab;
            return (
              <React.Fragment key={tab}>
                <button
                  className={cn(
                    'bucket-tab-btn py-1.5 px-5 rounded-full border border-[var(--border)] bg-none text-[13px] cursor-pointer text-[var(--muted)] transition-all hover:border-[var(--muted)] flex items-center gap-2',
                    isActive &&
                      'bg-[var(--text)] border-[var(--text)] text-[var(--surface)] font-medium'
                  )}
                  onClick={() => setActiveTab(tab)}
                >
                  {tab}
                  <span
                    className={cn(
                      'flex items-center justify-center min-w-[18px] h-[18px] px-1.5 rounded-full text-[10px] font-bold gap-1',
                      isActive
                        ? 'bg-[var(--muted)]/20 text-[var(--surface)]'
                        : 'bg-[var(--surface2)] text-[var(--muted)]'
                    )}
                  >
                    <span>{emoji}</span>
                    {count}
                  </span>
                </button>
                {(tab === 'Measures' || tab === 'Aggregators') && (
                  <div className="w-px h-4 bg-[var(--border)] mx-1" />
                )}
              </React.Fragment>
            );
          })}
        </div>
        <div className="bucket-actions">{renderActiveActionButton()}</div>
      </div>

      <div className="bucket-tab-content flex-1 overflow-hidden">
        {activeTab === 'Measures' && (
          <Measures
            categories={categories}
            taxonomies={taxonomies}
            measures={measures}
            onMeasuresChange={onMeasuresChange}
            onEditMeasure={(m, tab = 'general') => {
              setEditingMeasure(m);
              setInitialMeasureTab(tab);
              setIsModalOpen(true);
            }}
          />
        )}
        {activeTab === 'Spreaders' && <Spreaders />}
        {activeTab === 'Aggregators' && <Aggregators />}
        {activeTab === 'Converters' && (
          <Converters
            converters={converters}
            onEditConverter={c => {
              setEditingConverter(c);
              setIsConverterModalOpen(true);
            }}
          />
        )}
      </div>

      <MeasureModal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        onConfirm={handleMeasureConfirm}
        onDelete={handleMeasureDelete}
        existingMeasure={editingMeasure}
        categories={categories}
        initialTab={initialMeasureTab}
      />

      <ConverterModal
        isOpen={isConverterModalOpen}
        onClose={() => setIsConverterModalOpen(false)}
        onConfirm={handleConverterConfirm}
        onDelete={handleConverterDelete}
        existingConverter={editingConverter}
        measures={measures}
        categories={categories}
      />
    </div>
  );
};
