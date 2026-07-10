import React, { useEffect, useState } from 'react';
import { Plan } from '../../types';

interface PlanModalProps {
  isOpen: boolean;
  onClose: () => void;
  onConfirm: (data: Partial<Plan>) => void;
  existingPlan?: Plan | null;
}

export const PlanModal: React.FC<PlanModalProps> = ({
  isOpen,
  onClose,
  onConfirm,
  existingPlan,
}) => {
  const [label, setLabel] = useState('');
  const [key, setKey] = useState('');
  const [tooltip, setTooltip] = useState('');
  const [desc, setDesc] = useState('');
  const [kind, setKind] = useState<'report' | 'forecast'>('report');

  useEffect(() => {
    setTimeout(() => {
      if (existingPlan) {
        setLabel(existingPlan.label);
        setKey(existingPlan.key);
        setTooltip(existingPlan.tooltip);
        setDesc(existingPlan.desc);
        setKind(existingPlan.kind);
      } else {
        setLabel('');
        setKey('');
        setTooltip('');
        setDesc('');
        setKind('report');
      }
    }, 0);
  }, [existingPlan, isOpen]);

  if (!isOpen) return null;

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onConfirm({ label, key, tooltip, desc, kind });
  };

  return (
    <div className="modal-overlay fixed inset-0 bg-black/60 backdrop-blur-sm z-50 flex items-center justify-center p-4">
      <div className="modal-content bg-[var(--surface)] border border-[var(--border)] rounded-2xl w-full max-w-md shadow-2xl overflow-hidden animate-in fade-in zoom-in duration-200">
        <div className="p-6 border-b border-[var(--border)] flex items-center justify-between">
          <h2 className="text-[20px] font-semibold m-0">
            {existingPlan ? 'Edit Plan Type' : 'Add Plan Type'}
          </h2>
          <button
            onClick={onClose}
            className="p-2 hover:bg-[var(--surface2)] rounded-full border-none bg-none cursor-pointer transition-colors"
          >
            <svg
              width="20"
              height="20"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            >
              <line x1="18" y1="6" x2="6" y2="18"></line>
              <line x1="6" y1="6" x2="18" y2="18"></line>
            </svg>
          </button>
        </div>

        <form
          onSubmit={handleSubmit}
          className="p-6 flex flex-col gap-4 overflow-y-auto max-h-[70vh]"
        >
          <div className="flex flex-col gap-1.5">
            <label className="text-[11px] uppercase tracking-wider font-bold text-[var(--muted)]">
              Key
            </label>
            <input
              type="text"
              value={key}
              onChange={e => setKey(e.target.value)}
              placeholder="e.g. 13w_rev"
              className="bg-[var(--surface2)] border border-[var(--border)] rounded-lg p-2.5 text-[14px] text-[var(--text)] focus:outline-none focus:border-[var(--accent)] transition-colors"
              required
            />
          </div>

          <div className="flex flex-col gap-1.5">
            <label className="text-[11px] uppercase tracking-wider font-bold text-[var(--muted)]">
              Name
            </label>
            <input
              type="text"
              value={label}
              onChange={e => setLabel(e.target.value)}
              placeholder="e.g. 13-week Review"
              className="bg-[var(--surface2)] border border-[var(--border)] rounded-lg p-2.5 text-[14px] text-[var(--text)] focus:outline-none focus:border-[var(--accent)] transition-colors"
              required
            />
          </div>

          <div className="flex flex-col gap-1.5">
            <label className="text-[11px] uppercase tracking-wider font-bold text-[var(--muted)]">
              Tooltip
            </label>
            <input
              type="text"
              value={tooltip}
              onChange={e => setTooltip(e.target.value)}
              placeholder="Short summary..."
              className="bg-[var(--surface2)] border border-[var(--border)] rounded-lg p-2.5 text-[14px] text-[var(--text)] focus:outline-none focus:border-[var(--accent)] transition-colors"
            />
          </div>

          <div className="flex flex-col gap-1.5">
            <label className="text-[11px] uppercase tracking-wider font-bold text-[var(--muted)]">
              Description
            </label>
            <textarea
              value={desc}
              onChange={e => setDesc(e.target.value)}
              placeholder="Detailed description..."
              rows={3}
              className="bg-[var(--surface2)] border border-[var(--border)] rounded-lg p-2.5 text-[14px] text-[var(--text)] focus:outline-none focus:border-[var(--accent)] transition-colors resize-none"
            />
          </div>

          <div className="flex flex-col gap-1.5">
            <label className="text-[11px] uppercase tracking-wider font-bold text-[var(--muted)]">
              Plan Kind
            </label>
            <div className="grid grid-cols-2 gap-2">
              <button
                type="button"
                onClick={() => setKind('report')}
                className={`py-2 rounded-lg border text-[13px] font-medium transition-all ${
                  kind === 'report'
                    ? 'bg-[var(--accent)] border-[var(--accent)] text-white'
                    : 'bg-[var(--surface2)] border-[var(--border)] text-[var(--muted)] hover:border-[var(--muted)]'
                }`}
              >
                Report
              </button>
              <button
                type="button"
                onClick={() => setKind('forecast')}
                className={`py-2 rounded-lg border text-[13px] font-medium transition-all ${
                  kind === 'forecast'
                    ? 'bg-[var(--accent2)] border-[var(--accent2)] text-[#0f1117]'
                    : 'bg-[var(--surface2)] border-[var(--border)] text-[var(--muted)] hover:border-[var(--muted)]'
                }`}
              >
                Forecast
              </button>
            </div>
          </div>

          <div className="flex gap-3 mt-2">
            <button
              type="button"
              onClick={onClose}
              className="flex-1 py-2.5 rounded-xl border border-[var(--border)] bg-none text-[var(--text)] text-[14px] font-medium cursor-pointer hover:bg-[var(--surface2)] transition-colors"
            >
              Cancel
            </button>
            <button
              type="submit"
              className="flex-1 py-2.5 rounded-xl border-none bg-[var(--accent)] text-white text-[14px] font-semibold cursor-pointer hover:opacity-90 transition-opacity"
            >
              {existingPlan ? 'Save' : 'Create'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};
