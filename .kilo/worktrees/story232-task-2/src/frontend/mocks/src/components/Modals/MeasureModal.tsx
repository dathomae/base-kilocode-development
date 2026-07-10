import { X } from 'lucide-react';
import React, { useEffect, useState } from 'react';
import { cn } from '../../lib/utils';
import { Category, Measure } from '../../types';

interface MeasureModalProps {
  isOpen: boolean;
  onClose: () => void;
  onConfirm: (data: Measure) => void;
  onDelete?: (id: string) => void;
  existingMeasure?: Measure | null;
  categories: Category[];
  initialTab?: 'general' | 'locales' | 'properties';
}

export const MeasureModal: React.FC<MeasureModalProps> = ({
  isOpen,
  onClose,
  onConfirm,
  onDelete,
  existingMeasure,
  categories,
  initialTab = 'general',
}) => {
  const [key, setKey] = useState('');
  const [name, setName] = useState('');
  const [tooltip, setTooltip] = useState('');
  const [desc, setDesc] = useState('');
  const [dataType, setDataType] = useState('');
  const [units, setUnits] = useState('');
  const [activeTab, setActiveTab] = useState<'general' | 'locales' | 'properties'>(initialTab);

  useEffect(() => {
    if (isOpen) {
      setTimeout(() => {
        if (existingMeasure) {
          setKey(existingMeasure.key);
          setName(existingMeasure.name);
          setTooltip(existingMeasure.tooltip);
          setDesc(existingMeasure.desc);
          setDataType(existingMeasure.dataType);
          setUnits(existingMeasure.units);
        } else {
          setKey('');
          setName('');
          setTooltip('');
          setDesc('');
          setDataType('');
          setUnits('');
        }
        setActiveTab(initialTab);
      }, 0);
    }
  }, [existingMeasure, isOpen, initialTab]);

  if (!isOpen) return null;

  const dataTypesCat = categories.find(c => c.key === 'data-types');
  const dataTypes = dataTypesCat ? dataTypesCat._children : [];

  const unitCategories = categories.filter(
    c => c.isBaseUnit || c.key === 'channel' || c.key === 'product_class'
  );

  const handleConfirm = () => {
    onConfirm({
      id: existingMeasure?.id || `m-${key}`,
      key,
      name,
      tooltip,
      desc,
      dataType,
      units,
      _locales: existingMeasure?._locales || [],
      _props: existingMeasure?._props || [],
    });
  };

  const isNameValid = name.length <= 64;
  const isTooltipValid = tooltip.length >= 16 && tooltip.length <= 256;
  const isDescValid = desc.length > 40;
  const isValid = key && name && isNameValid && isTooltipValid && isDescValid && dataType && units;

  return (
    <div className="fixed inset-0 z-[1000] flex items-center justify-center bg-black/50 backdrop-blur-sm animate-in fade-in duration-200">
      <div className="bg-[var(--surface)] border border-[var(--border)] rounded-xl shadow-2xl w-full max-w-2xl overflow-hidden flex flex-col max-h-[90vh]">
        <div className="p-5 border-b border-[var(--border)] flex items-center justify-between bg-[var(--surface2)]/50">
          <h2 className="text-xl font-semibold tracking-tight">
            {existingMeasure ? 'Edit Measure' : 'Add Measure'}
          </h2>
          <button
            onClick={onClose}
            className="p-1.5 rounded-lg hover:bg-[var(--surface2)] text-[var(--muted)] transition-colors"
          >
            <X size={20} />
          </button>
        </div>

        <div className="flex border-b border-[var(--border)] bg-[var(--surface2)]/30">
          {(['general', 'locales', 'properties'] as const).map(tab => (
            <button
              key={tab}
              onClick={() => setActiveTab(tab)}
              className={cn(
                'px-6 py-3 text-sm font-medium transition-all border-b-2 border-transparent hover:text-[var(--text)] capitalize',
                activeTab === tab
                  ? 'border-[var(--accent)] text-[var(--accent)] bg-[var(--surface)]'
                  : 'text-[var(--muted)]'
              )}
            >
              {tab}
            </button>
          ))}
        </div>

        <div className="flex-1 overflow-y-auto p-8 space-y-6">
          {activeTab === 'general' && (
            <>
              <div className="grid grid-cols-2 gap-6">
                <div className="space-y-2">
                  <label className="text-xs font-bold uppercase tracking-wider text-[var(--muted)]">
                    Key
                  </label>
                  <input
                    value={key}
                    onChange={e => setKey(e.target.value)}
                    className="w-full bg-[var(--surface2)] border border-[var(--border)] rounded-lg px-4 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-[var(--accent)]/50 transition-all font-mono"
                    placeholder="e.g. revenue"
                  />
                </div>
                <div className="space-y-2">
                  <label className="text-xs font-bold uppercase tracking-wider text-[var(--muted)]">
                    Name
                  </label>
                  <input
                    value={name}
                    onChange={e => setName(e.target.value)}
                    className={cn(
                      'w-full bg-[var(--surface2)] border border-[var(--border)] rounded-lg px-4 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-[var(--accent)]/50 transition-all',
                      !isNameValid && 'border-red-500/50 focus:ring-red-500/50'
                    )}
                    placeholder="e.g. Total Revenue"
                  />
                  {!isNameValid && <p className="text-[10px] text-red-500">Max 64 characters</p>}
                </div>
              </div>

              <div className="space-y-2">
                <label className="text-xs font-bold uppercase tracking-wider text-[var(--muted)]">
                  Tooltip
                </label>
                <input
                  value={tooltip}
                  onChange={e => setTooltip(e.target.value)}
                  className={cn(
                    'w-full bg-[var(--surface2)] border border-[var(--border)] rounded-lg px-4 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-[var(--accent)]/50 transition-all',
                    tooltip.length > 0 &&
                      !isTooltipValid &&
                      'border-red-500/50 focus:ring-red-500/50'
                  )}
                  placeholder="Short description for tooltips (16-256 chars)"
                />
                {tooltip.length > 0 && !isTooltipValid && (
                  <p className="text-[10px] text-red-500">
                    Must be between 16 and 256 characters (Current: {tooltip.length})
                  </p>
                )}
              </div>

              <div className="space-y-2">
                <label className="text-xs font-bold uppercase tracking-wider text-[var(--muted)]">
                  Description
                </label>
                <textarea
                  value={desc}
                  onChange={e => setDesc(e.target.value)}
                  rows={4}
                  className={cn(
                    'w-full bg-[var(--surface2)] border border-[var(--border)] rounded-lg px-4 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-[var(--accent)]/50 transition-all resize-none',
                    desc.length > 0 && !isDescValid && 'border-red-500/50 focus:ring-red-500/50'
                  )}
                  placeholder="Full detailed description (> 40 chars)"
                />
                {desc.length > 0 && !isDescValid && (
                  <p className="text-[10px] text-red-500">
                    Must be more than 40 characters (Current: {desc.length})
                  </p>
                )}
              </div>

              <div className="grid grid-cols-2 gap-6">
                <div className="space-y-2">
                  <label className="text-xs font-bold uppercase tracking-wider text-[var(--muted)]">
                    Data Type
                  </label>
                  <select
                    value={dataType}
                    onChange={e => setDataType(e.target.value)}
                    className="w-full bg-[var(--surface2)] border border-[var(--border)] rounded-lg px-4 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-[var(--accent)]/50 transition-all"
                  >
                    <option value="">Select Data Type</option>
                    {dataTypes.map(dt => (
                      <option key={dt.id} value={dt.key}>
                        {dt.name}
                      </option>
                    ))}
                  </select>
                </div>
                <div className="space-y-3">
                  <div className="flex items-center justify-between">
                    <label className="text-xs font-bold uppercase tracking-wider text-[var(--muted)]">
                      Units
                    </label>
                    <div className="flex gap-2">
                      {['*', '/', '^'].map(op => (
                        <button
                          key={op}
                          onClick={() =>
                            setUnits(
                              prev => prev + (prev && !prev.endsWith(' ') ? ' ' : '') + op + ' '
                            )
                          }
                          className="px-2 py-0.5 bg-[var(--surface2)] border border-[var(--border)] rounded text-[var(--warn)] font-mono text-xs hover:bg-[var(--border)] transition-colors"
                        >
                          {op}
                        </button>
                      ))}
                    </div>
                  </div>
                  <div className="relative">
                    <input
                      value={units}
                      onChange={e => setUnits(e.target.value)}
                      className="w-full bg-[var(--surface2)] border border-[var(--border)] rounded-lg px-4 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-[var(--accent)]/50 transition-all font-mono"
                      placeholder="e.g. distance / time ^ 2"
                    />
                    <div className="mt-2 flex flex-wrap gap-1.5 max-h-32 overflow-y-auto p-1">
                      {unitCategories.map(cat => (
                        <React.Fragment key={cat.id}>
                          {cat._children.map(el => (
                            <button
                              key={el.id}
                              onClick={() =>
                                setUnits(
                                  prev =>
                                    prev + (prev && !prev.endsWith(' ') ? ' ' : '') + el.key + ' '
                                )
                              }
                              className="px-2 py-1 bg-[var(--surface2)] border border-[var(--border)] rounded-md text-[10px] text-[var(--accent2)] hover:border-[var(--accent2)] transition-all whitespace-nowrap"
                              title={el.name}
                            >
                              {el.key}
                            </button>
                          ))}
                        </React.Fragment>
                      ))}
                    </div>
                  </div>
                  <div className="p-3 bg-[var(--surface2)]/50 border border-[var(--border)] rounded-lg">
                    <p className="text-[10px] uppercase font-bold text-[var(--muted)] mb-2 tracking-widest">
                      Preview
                    </p>
                    <div className="font-mono text-sm flex flex-wrap gap-1 items-center">
                      {units.split(/(\s+|\*|\/|\^)/).map((part, i) => {
                        if (['*', '/', '^'].includes(part.trim())) {
                          return (
                            <span key={i} style={{ color: 'var(--warn)' }}>
                              {part}
                            </span>
                          );
                        }
                        if (part.trim() && !['*', '/', '^'].includes(part.trim())) {
                          return (
                            <span key={i} className="text-[var(--accent2)]">
                              {part}
                            </span>
                          );
                        }
                        return <span key={i}>{part}</span>;
                      })}
                      {!units && (
                        <span className="text-[var(--muted)] italic text-xs">No units defined</span>
                      )}
                    </div>
                  </div>
                </div>
              </div>
            </>
          )}

          {activeTab === 'locales' && (
            <div className="text-center py-10 text-[var(--muted)] italic">
              Localization management coming soon.
            </div>
          )}

          {activeTab === 'properties' && (
            <div className="text-center py-10 text-[var(--muted)] italic">
              Custom properties management coming soon.
            </div>
          )}
        </div>

        <div className="p-6 border-t border-[var(--border)] bg-[var(--surface2)]/50 flex items-center justify-between">
          <div>
            {existingMeasure && onDelete && (
              <button
                onClick={() => onDelete(existingMeasure.id)}
                className="px-5 py-2 rounded-lg border border-red-500/20 text-red-500 hover:bg-red-500/10 transition-all text-sm font-medium"
              >
                Delete Measure
              </button>
            )}
          </div>
          <div className="flex gap-3">
            <button
              onClick={onClose}
              className="px-5 py-2 rounded-lg border border-[var(--border)] text-[var(--muted)] hover:bg-[var(--surface2)] transition-all text-sm font-medium"
            >
              Cancel
            </button>
            <button
              onClick={handleConfirm}
              disabled={!isValid}
              className={cn(
                'px-6 py-2 rounded-lg bg-[var(--text)] text-[var(--surface)] font-medium text-sm transition-all shadow-lg shadow-black/10',
                !isValid && 'opacity-50 cursor-not-allowed'
              )}
            >
              {existingMeasure ? 'Save Changes' : 'Add Measure'}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};
