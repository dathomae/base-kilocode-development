import React, { useEffect, useState } from 'react';
import { cn } from '../../lib/utils';
import { Metric } from '../../types';

interface MetricModalProps {
  isOpen: boolean;
  onClose: () => void;
  onConfirm: (metric: Metric) => void;
  onDelete?: (id: number) => void;
  existingMetric?: Metric | null;
}

const ALL_DIMS = [
  'country',
  'region',
  'metro',
  'warehouse',
  'category',
  'product',
  'channel',
  'customer_segment',
  'supplier',
  'sku',
];

export const MetricModal: React.FC<MetricModalProps> = ({
  isOpen,
  onClose,
  onConfirm,
  onDelete,
  existingMetric,
}) => {
  const [measure, setMeasure] = useState('inventory');
  const [grain, setGrain] = useState<'daily' | 'weekly' | 'monthly' | 'quarterly'>('daily');
  const [selectedDims, setSelectedDims] = useState<string[]>([]);
  const [hiAvail, setHiAvail] = useState<string | null>(null);
  const [hiChosen, setHiChosen] = useState<string | null>(null);

  useEffect(() => {
    setTimeout(() => {
      if (existingMetric) {
        setMeasure(existingMetric.measure);
        setGrain(existingMetric.grain);
        setSelectedDims(existingMetric.dims);
      } else {
        setMeasure('inventory');
        setGrain('daily');
        setSelectedDims([]);
      }
      setHiAvail(null);
      setHiChosen(null);
    }, 0);
  }, [existingMetric, isOpen]);

  if (!isOpen) return null;

  const dimRight = () => {
    if (!hiAvail) return;
    setSelectedDims([...selectedDims, hiAvail]);
    setHiChosen(hiAvail);
    setHiAvail(null);
  };

  const dimLeft = () => {
    if (!hiChosen) return;
    setSelectedDims(selectedDims.filter(d => d !== hiChosen));
    setHiAvail(hiChosen);
    setHiChosen(null);
  };

  const moveUp = () => {
    if (!hiChosen) return;
    const i = selectedDims.indexOf(hiChosen);
    if (i > 0) {
      const next = [...selectedDims];
      [next[i - 1], next[i]] = [next[i], next[i - 1]];
      setSelectedDims(next);
    }
  };

  const moveDown = () => {
    if (!hiChosen) return;
    const i = selectedDims.indexOf(hiChosen);
    if (i < selectedDims.length - 1) {
      const next = [...selectedDims];
      [next[i + 1], next[i]] = [next[i], next[i + 1]];
      setSelectedDims(next);
    }
  };

  return (
    <div className="modal-overlay open" onClick={e => e.target === e.currentTarget && onClose()}>
      <div className="modal">
        <div className="modal-header p-5 pb-0 shrink-0">
          <h3 className="text-base mb-3.5">
            {existingMetric ? 'Edit Metric Declaration' : 'Add Metric Declaration'}
          </h3>
        </div>
        <div className="modal-body flex-1 overflow-y-auto p-5">
          <div className="form-row mb-3.5">
            <label className="block text-[11px] text-[var(--muted)] uppercase tracking-[0.5px] mb-1.5">
              Measure
            </label>
            <select
              className="form-input w-full p-2 rounded-lg bg-[var(--surface)] border border-[var(--border)] text-[var(--text)] text-[13px] cursor-pointer"
              value={measure}
              onChange={e => setMeasure(e.target.value)}
            >
              <option>inventory</option>
              <option>demand</option>
              <option>supply</option>
              <option>revenue</option>
              <option>cost</option>
              <option>returns</option>
              <option>capacity</option>
              <option>fulfillment</option>
            </select>
          </div>
          <div className="form-row mb-3.5">
            <label className="block text-[11px] text-[var(--muted)] uppercase tracking-[0.5px] mb-1.5">
              Dimensions{' '}
              <span className="text-[10px] lowercase tracking-normal font-normal">
                — double-click to move; ▲▼ to reorder
              </span>
            </label>
            <div className="dim-builder flex gap-2 items-start">
              <div className="dim-col flex flex-col flex-1 gap-1.5">
                <div className="dim-col-label text-[10px] text-[var(--muted)] uppercase tracking-[0.5px]">
                  Available
                </div>
                <div className="dim-available bg-[var(--surface)] border border-[var(--border)] rounded-lg min-h-[148px] max-h-[148px] overflow-y-auto p-1">
                  {ALL_DIMS.filter(d => !selectedDims.includes(d)).map(d => (
                    <div
                      key={d}
                      className={cn(
                        'dim-item p-1 px-2 rounded-sm text-[12px] cursor-pointer text-[var(--muted)] transition-colors hover:bg-[#6c63ff]/15 hover:text-[var(--text)]',
                        hiAvail === d && 'bg-[#6c63ff]/30 text-[var(--text)]'
                      )}
                      onClick={() => {
                        setHiAvail(d);
                        setHiChosen(null);
                      }}
                      onDoubleClick={() => {
                        setHiAvail(d);
                        dimRight();
                      }}
                    >
                      {d}
                    </div>
                  ))}
                </div>
              </div>
              <div className="dim-arrows flex flex-col justify-center gap-1.5 pt-5">
                <button
                  className="dim-move-btn p-1 px-2 rounded-md bg-[var(--surface)] border border-[var(--border)] text-[var(--muted)] text-[11px] cursor-pointer hover:border-[var(--accent)] hover:text-[var(--text)]"
                  onClick={dimRight}
                >
                  ▶
                </button>
                <button
                  className="dim-move-btn p-1 px-2 rounded-md bg-[var(--surface)] border border-[var(--border)] text-[var(--muted)] text-[11px] cursor-pointer hover:border-[var(--accent)] hover:text-[var(--text)]"
                  onClick={dimLeft}
                >
                  ◀
                </button>
              </div>
              <div className="dim-col flex flex-col flex-1 gap-1.5">
                <div className="dim-col-label text-[10px] text-[var(--muted)] uppercase tracking-[0.5px]">
                  Selected
                </div>
                <div className="dim-selected bg-[var(--surface)] border border-[var(--border)] rounded-lg min-h-[148px] max-h-[148px] overflow-y-auto p-1">
                  {selectedDims.map(d => (
                    <div
                      key={d}
                      className={cn(
                        'dim-item p-1 px-2 rounded-sm text-[12px] cursor-pointer text-[var(--muted)] transition-colors hover:bg-[#6c63ff]/15 hover:text-[var(--text)]',
                        hiChosen === d && 'bg-[#6c63ff]/30 text-[var(--text)]'
                      )}
                      onClick={() => {
                        setHiChosen(d);
                        setHiAvail(null);
                      }}
                      onDoubleClick={() => {
                        setHiChosen(d);
                        dimLeft();
                      }}
                    >
                      {d}
                    </div>
                  ))}
                </div>
                <div className="dim-order-btns flex gap-1.5">
                  <button
                    className="dim-move-btn flex-1 p-1 px-2 rounded-md bg-[var(--surface)] border border-[var(--border)] text-[var(--muted)] text-[11px] cursor-pointer hover:border-[var(--accent)] hover:text-[var(--text)]"
                    onClick={moveUp}
                  >
                    ▲
                  </button>
                  <button
                    className="dim-move-btn flex-1 p-1 px-2 rounded-md bg-[var(--surface)] border border-[var(--border)] text-[var(--muted)] text-[11px] cursor-pointer hover:border-[var(--accent)] hover:text-[var(--text)]"
                    onClick={moveDown}
                  >
                    ▼
                  </button>
                </div>
              </div>
            </div>
          </div>
          <div className="form-row mb-3.5">
            <label className="block text-[11px] text-[var(--muted)] uppercase tracking-[0.5px] mb-1.5">
              Grain
            </label>
            <select
              className="form-input w-full p-2 rounded-lg bg-[var(--surface)] border border-[var(--border)] text-[var(--text)] text-[13px] cursor-pointer"
              value={grain}
              onChange={e =>
                setGrain(e.target.value as 'daily' | 'weekly' | 'monthly' | 'quarterly')
              }
            >
              <option>daily</option>
              <option>weekly</option>
              <option>monthly</option>
              <option>quarterly</option>
            </select>
          </div>
        </div>
        <div className="modal-footer p-3.5 px-6 border-t border-[var(--border)] flex items-center gap-2.5 shrink-0">
          {existingMetric && onDelete && (
            <button
              className="btn-delete-modal p-1.5 px-3.5 rounded-lg bg-none border border-[#5a2020] text-[#a05050] text-[12px] cursor-pointer hover:bg-[#3a1515] hover:border-[var(--danger)] hover:text-[var(--danger)]"
              onClick={() => onDelete(existingMetric.id)}
            >
              🗑 Delete
            </button>
          )}
          <span className="flex-1"></span>
          <button
            className="btn-cancel p-1.5 px-4 rounded-lg bg-none border border-[var(--border)] text-[var(--muted)] text-[13px] cursor-pointer"
            onClick={onClose}
          >
            Cancel
          </button>
          <button
            className="btn-confirm p-1.5 px-4 rounded-lg bg-[var(--accent)] border-none text-white text-[13px] cursor-pointer"
            onClick={() =>
              onConfirm({
                id: existingMetric?.id || Date.now(),
                measure,
                dims: selectedDims,
                grain,
              })
            }
          >
            {existingMetric ? 'Save Changes' : 'Add Metric'}
          </button>
        </div>
      </div>
    </div>
  );
};
