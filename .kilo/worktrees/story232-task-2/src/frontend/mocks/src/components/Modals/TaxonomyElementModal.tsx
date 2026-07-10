import React, { useState } from 'react';
import { Category } from '../../types';

interface TaxonomyElementModalProps {
  isOpen: boolean;
  onClose: () => void;
  onConfirm: (elKey: string) => void;
  allowedCats: string[];
  categories: Category[];
}

export const TaxonomyElementModal: React.FC<TaxonomyElementModalProps> = ({
  isOpen,
  onClose,
  onConfirm,
  allowedCats,
  categories,
}) => {
  const [selectedElKey, setSelectedElKey] = useState('');

  if (!isOpen) return null;

  const filteredCategories = categories.filter(c => allowedCats.includes(c.key));
  const allElements = filteredCategories.flatMap(c => c._children);

  return (
    <div className="modal-overlay fixed inset-0 bg-black/50 flex items-center justify-center z-[1000] backdrop-blur-sm">
      <div className="modal-content bg-[var(--surface)] w-[400px] rounded-2xl shadow-2xl border border-[var(--border)] overflow-hidden flex flex-col max-h-[80vh]">
        <div className="modal-header p-5 border-b border-[var(--border)] flex justify-between items-center bg-[var(--surface2)]">
          <h3 className="text-[16px] font-semibold text-[var(--text)]">Add Element to Taxonomy</h3>
          <button
            className="close-btn bg-none border-none text-[20px] cursor-pointer text-[var(--muted)] hover:text-[var(--text)]"
            onClick={onClose}
          >
            &times;
          </button>
        </div>

        <div className="modal-body p-6 overflow-y-auto">
          <div className="form-group mb-4">
            <label className="block text-[12px] font-medium text-[var(--muted)] mb-1.5 uppercase tracking-wider">
              Select Element
            </label>
            <select
              className="w-full p-2.5 rounded-lg bg-[var(--surface2)] border border-[var(--border)] text-[var(--text)] text-[14px] outline-none focus:border-[var(--accent)] transition-all"
              value={selectedElKey}
              onChange={e => setSelectedElKey(e.target.value)}
            >
              <option value="">— Choose an element —</option>
              {allElements.map(el => (
                <option key={el.key} value={el.key}>
                  {el.name} ({el.key})
                </option>
              ))}
            </select>
          </div>
        </div>

        <div className="modal-footer p-4 border-t border-[var(--border)] flex justify-end gap-3 bg-[var(--surface2)]">
          <button
            className="btn-cancel p-1.5 px-4 rounded-lg bg-none border border-[var(--border)] text-[var(--muted)] text-[13px] cursor-pointer"
            onClick={onClose}
          >
            Cancel
          </button>
          <button
            className="btn-confirm p-1.5 px-4 rounded-lg bg-[var(--accent)] border-none text-white text-[13px] cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
            onClick={() => onConfirm(selectedElKey)}
            disabled={!selectedElKey}
          >
            Add Element
          </button>
        </div>
      </div>
    </div>
  );
};
