import React, { useEffect, useState } from 'react';
import { cn } from '../../lib/utils';
import { Category, CategoryElement, Locale, Property } from '../../types';
// import JSONEditor from 'jsoneditor'; // Removed as it was unused
// import 'jsoneditor/dist/jsoneditor.min.css';

interface CategoryModalProps {
  isOpen: boolean;
  onClose: () => void;
  onConfirm: (data: Partial<Category>) => void;
  onDelete?: () => void;

  mode: 'addCat' | 'editCat' | 'addEl' | 'editEl';
  existingData?: Category | CategoryElement | null;
  initialTab?: 'general' | 'properties' | 'locales';
}

export const CategoryModal: React.FC<CategoryModalProps> = ({
  isOpen,
  onClose,
  onConfirm,
  onDelete,
  mode,
  existingData,
  initialTab,
}) => {
  const [activeTab, setActiveTab] = useState<'general' | 'properties' | 'locales'>('general');
  const [key, setKey] = useState('');
  const [name, setName] = useState('');
  const [tooltip, setTooltip] = useState('');
  const [desc, setDesc] = useState('');
  const [isBaseUnit, setIsBaseUnit] = useState(false);
  const [locales, setLocales] = useState<Locale[]>([]);
  const [props, setProps] = useState<Property[]>([]);

  useEffect(() => {
    if (!isOpen) return;

    setTimeout(() => {
      if (existingData) {
        setKey(existingData.key);
        setName(existingData.name);
        setTooltip(existingData.tooltip);
        setDesc(existingData.desc);
        setIsBaseUnit(existingData.isBaseUnit || false);
        setLocales(JSON.parse(JSON.stringify(existingData._locales || [])));
        setProps(JSON.parse(JSON.stringify(existingData._props || [])));
      } else {
        setKey('');
        setName('');
        setTooltip('');
        setDesc('');
        setIsBaseUnit(false);
        setLocales([{ lang: 'en', region: '', name: '', tooltip: '', desc: '' }]);
        setProps([]);
      }
      setActiveTab(initialTab || 'general');
    }, 0);
  }, [existingData, isOpen, initialTab]);

  if (!isOpen) return null;

  const handleConfirm = () => {
    onConfirm({ key, name, tooltip, desc, isBaseUnit, _locales: locales, _props: props });
  };

  return (
    <div className="modal-overlay open" onClick={e => e.target === e.currentTarget && onClose()}>
      <div className="modal">
        <div className="modal-header p-5 pb-0 shrink-0">
          <h3 className="text-base mb-3.5">
            {mode.startsWith('add') ? 'Add' : 'Edit'}{' '}
            {mode.includes('Cat') ? 'Category' : 'Element'}
          </h3>
          <div className="modal-tabs flex gap-0.5 border-b border-[var(--border)] px-6">
            <button
              className={cn(
                'modal-tab p-2 px-4 bg-none border-none cursor-pointer text-[13px] text-[var(--muted)] border-b-2 border-transparent -mb-px transition-colors hover:text-[var(--text)]',
                activeTab === 'general' &&
                  'text-[var(--accent2)] border-b-[var(--accent2)] font-semibold'
              )}
              onClick={() => setActiveTab('general')}
            >
              General
            </button>
            <button
              className={cn(
                'modal-tab p-2 px-4 bg-none border-none cursor-pointer text-[13px] text-[var(--muted)] border-b-2 border-transparent -mb-px transition-colors hover:text-[var(--text)]',
                activeTab === 'locales' &&
                  'text-[var(--accent2)] border-b-[var(--accent2)] font-semibold'
              )}
              onClick={() => setActiveTab('locales')}
            >
              Locales
            </button>
            <button
              className={cn(
                'modal-tab p-2 px-4 bg-none border-none cursor-pointer text-[13px] text-[var(--muted)] border-b-2 border-transparent -mb-px transition-colors hover:text-[var(--text)]',
                activeTab === 'properties' &&
                  'text-[var(--accent2)] border-b-[var(--accent2)] font-semibold'
              )}
              onClick={() => setActiveTab('properties')}
            >
              Properties
            </button>
          </div>
        </div>
        <div className="modal-body flex-1 overflow-y-auto p-5">
          {activeTab === 'general' ? (
            <div>
              <div className="form-row mb-3.5">
                <label className="block text-[11px] text-[var(--muted)] uppercase tracking-[0.5px] mb-1.5">
                  Key
                </label>
                <input
                  className="form-input w-full p-2 rounded-lg bg-[var(--surface)] border border-[var(--border)] text-[var(--text)] text-[13px] font-mono"
                  value={key}
                  onChange={e => setKey(e.target.value)}
                  disabled={mode.startsWith('edit')}
                  placeholder="e.g. product_class"
                />
              </div>
              <div className="form-row mb-3.5">
                <label className="block text-[11px] text-[var(--muted)] uppercase tracking-[0.5px] mb-1.5">
                  Name
                </label>
                <input
                  className="form-input w-full p-2 rounded-lg bg-[var(--surface)] border border-[var(--border)] text-[var(--text)] text-[13px]"
                  value={name}
                  onChange={e => setName(e.target.value)}
                  placeholder="Display name"
                />
              </div>
              <div className="form-row mb-3.5">
                <label className="block text-[11px] text-[var(--muted)] uppercase tracking-[0.5px] mb-1.5">
                  Tooltip
                </label>
                <input
                  className="form-input w-full p-2 rounded-lg bg-[var(--surface)] border border-[var(--border)] text-[var(--text)] text-[13px]"
                  value={tooltip}
                  onChange={e => setTooltip(e.target.value)}
                  placeholder="Short tooltip text"
                />
              </div>
              <div className="form-row mb-3.5">
                <label className="block text-[11px] text-[var(--muted)] uppercase tracking-[0.5px] mb-1.5">
                  Description
                </label>
                <textarea
                  className="form-input w-full p-2 rounded-lg bg-[var(--surface)] border border-[var(--border)] text-[var(--text)] text-[13px] resize-vertical"
                  rows={3}
                  value={desc}
                  onChange={e => setDesc(e.target.value)}
                  placeholder="Longer description..."
                />
              </div>
              <div className="form-row mb-3.5 flex items-center gap-2">
                <input
                  type="checkbox"
                  id="isBaseUnit"
                  className="w-4 h-4 accent-[var(--accent)] cursor-pointer"
                  checked={isBaseUnit}
                  onChange={e => setIsBaseUnit(e.target.checked)}
                />
                <label
                  htmlFor="isBaseUnit"
                  className="text-[11px] text-[var(--muted)] uppercase tracking-[0.5px] cursor-pointer select-none"
                >
                  Base Unit
                </label>
              </div>
            </div>
          ) : activeTab === 'locales' ? (
            <div className="locale-list flex flex-col gap-3">
              {locales.map((loc, i) => (
                <div
                  key={i}
                  className="locale-row bg-[var(--surface)] border border-[var(--border)] rounded-xl p-3"
                >
                  <div className="flex items-center gap-2 mb-3">
                    <input
                      className="flex-1 p-1.5 rounded bg-[var(--surface2)] border border-[var(--border)] text-[var(--text)] text-[12px]"
                      value={loc.lang}
                      onChange={e => {
                        const next = [...locales];
                        next[i].lang = e.target.value;
                        setLocales(next);
                      }}
                      placeholder="Language (e.g. en)"
                    />
                    <input
                      className="flex-1 p-1.5 rounded bg-[var(--surface2)] border border-[var(--border)] text-[var(--text)] text-[12px]"
                      value={loc.region}
                      onChange={e => {
                        const next = [...locales];
                        next[i].region = e.target.value;
                        setLocales(next);
                      }}
                      placeholder="Region (optional)"
                    />
                    <button
                      className="p-1 px-2 text-[var(--muted)] hover:text-[var(--danger)] cursor-pointer"
                      onClick={() => setLocales(locales.filter((_, idx) => idx !== i))}
                    >
                      ✕
                    </button>
                  </div>
                  <div className="grid grid-cols-2 gap-2">
                    <input
                      className="p-1.5 rounded bg-[var(--surface2)] border border-[var(--border)] text-[var(--text)] text-[12px]"
                      value={loc.name}
                      onChange={e => {
                        const next = [...locales];
                        next[i].name = e.target.value;
                        setLocales(next);
                      }}
                      placeholder="Localized Name"
                    />
                    <input
                      className="p-1.5 rounded bg-[var(--surface2)] border border-[var(--border)] text-[var(--text)] text-[12px]"
                      value={loc.tooltip}
                      onChange={e => {
                        const next = [...locales];
                        next[i].tooltip = e.target.value;
                        setLocales(next);
                      }}
                      placeholder="Localized Tooltip"
                    />
                  </div>
                </div>
              ))}
              <button
                className="btn-add-panel inline-flex items-center gap-1.5 p-1 px-3 rounded-md bg-[var(--accent)] border-none text-white text-[12px] cursor-pointer hover:opacity-85 mt-2 self-start"
                onClick={() =>
                  setLocales([
                    ...locales,
                    { lang: '', region: '', name: '', tooltip: '', desc: '' },
                  ])
                }
              >
                ＋ Add Locale
              </button>
            </div>
          ) : (
            <div className="prop-list flex flex-col gap-2">
              {props.map((p, i) => (
                <div
                  key={p.id}
                  className="prop-row bg-[var(--surface)] border border-[var(--border)] rounded-xl overflow-hidden"
                >
                  <div className="prop-row-head flex items-center gap-2 p-2">
                    <input
                      className="prop-key-input flex-1 font-mono text-[12px] bg-transparent border-none border-b border-[var(--border)] text-[var(--accent2)] p-0.5 outline-none focus:border-b-[var(--accent)]"
                      value={p.key}
                      onChange={e => {
                        const next = [...props];
                        next[i].key = e.target.value;
                        setProps(next);
                      }}
                    />
                    <select
                      className="prop-type-select text-[11px] bg-[var(--surface2)] border border-[var(--border)] text-[var(--muted)] rounded-md p-1 cursor-pointer"
                      value={p.type}
                      onChange={e => {
                        const next = [...props];
                        next[i].type = e.target.value as Property['type'];

                        setProps(next);
                      }}
                    >
                      <option>string</option>
                      <option>number</option>
                      <option>boolean</option>
                      <option>object</option>
                      <option>array</option>
                    </select>
                    <button
                      className="prop-del-btn bg-none border-none cursor-pointer text-[var(--muted)] text-[14px] p-1 shrink-0 hover:text-[var(--danger)]"
                      onClick={() => setProps(props.filter((_, idx) => idx !== i))}
                    >
                      ✕
                    </button>
                  </div>
                </div>
              ))}
              <button
                className="btn-add-panel inline-flex items-center gap-1.5 p-1 px-3 rounded-md bg-[var(--accent)] border-none text-white text-[12px] cursor-pointer hover:opacity-85 mt-3"
                onClick={() =>
                  setProps([
                    ...props,
                    { id: 'p' + Date.now(), key: 'new_prop', type: 'string', value: '' },
                  ])
                }
              >
                ＋ Add Property
              </button>
            </div>
          )}
        </div>
        <div className="modal-footer p-3.5 px-6 border-t border-[var(--border)] flex items-center gap-2.5 shrink-0">
          {mode.startsWith('edit') && onDelete && (
            <button
              className="btn-delete-modal p-1.5 px-3.5 rounded-lg bg-none border border-[#5a2020] text-[#a05050] text-[12px] cursor-pointer hover:bg-[#3a1515] hover:border-[var(--danger)] hover:text-[var(--danger)]"
              onClick={onDelete}
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
            onClick={handleConfirm}
          >
            {mode.startsWith('edit') ? 'Save Changes' : 'Add'}
          </button>
        </div>
      </div>
    </div>
  );
};
