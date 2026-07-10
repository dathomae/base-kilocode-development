import React, { useEffect, useState } from 'react';
import { cn } from '../../lib/utils';
import { Category, Locale, Property, Taxonomy, TaxonomyNode } from '../../types';

interface TaxonomyModalProps {
  isOpen: boolean;
  onClose: () => void;
  onConfirm: (data: Partial<Taxonomy>) => void;

  onDelete?: () => void;
  existingTaxonomy?: Taxonomy | null;
  categories: Category[];
  initialTab?: 'general' | 'properties' | 'locales' | 'dag';
}

export const TaxonomyModal: React.FC<TaxonomyModalProps> = ({
  isOpen,
  onClose,
  onConfirm,
  onDelete,
  existingTaxonomy,
  categories,
  initialTab,
}) => {
  const [activeTab, setActiveTab] = useState<'general' | 'properties' | 'locales' | 'dag'>(
    'general'
  );
  const [key, setKey] = useState('');
  const [name, setName] = useState('');
  const [tooltip, setTooltip] = useState('');
  const [desc, setDesc] = useState('');
  const [locales, setLocales] = useState<Locale[]>([]);
  const [props, setProps] = useState<Property[]>([]);
  const [allowedCats, setAllowedCats] = useState<string[]>([]);
  const [dag, setDag] = useState<TaxonomyNode[]>([]);

  useEffect(() => {
    setTimeout(() => {
      if (existingTaxonomy) {
        setKey(existingTaxonomy.key);
        setName(existingTaxonomy.name);
        setTooltip(existingTaxonomy.tooltip);
        setDesc(existingTaxonomy.desc);
        setLocales(JSON.parse(JSON.stringify(existingTaxonomy._locales || [])));
        setProps(JSON.parse(JSON.stringify(existingTaxonomy._props || [])));
        setAllowedCats(existingTaxonomy.allowedCats.slice());
        setDag(JSON.parse(JSON.stringify(existingTaxonomy.dag || [])));
      } else {
        setKey('');
        setName('');
        setTooltip('');
        setDesc('');
        setLocales([{ lang: 'en', region: '', name: '', tooltip: '', desc: '' }]);
        setProps([]);
        setAllowedCats([]);
        setDag([]);
      }
      setActiveTab(initialTab || 'general');
    }, 0);
  }, [existingTaxonomy, isOpen, initialTab]);

  if (!isOpen) return null;

  const handleConfirm = () => {
    onConfirm({ key, name, tooltip, desc, _locales: locales, _props: props, allowedCats, dag });
  };

  return (
    <div className="modal-overlay open" onClick={e => e.target === e.currentTarget && onClose()}>
      <div className="modal w-[620px]">
        <div className="modal-header p-5 pb-0 shrink-0">
          <h3 className="text-base mb-3.5">
            {existingTaxonomy ? 'Edit Taxonomy' : 'Add Taxonomy'}
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
            <button
              className={cn(
                'modal-tab p-2 px-4 bg-none border-none cursor-pointer text-[13px] text-[var(--muted)] border-b-2 border-transparent -mb-px transition-colors hover:text-[var(--text)]',
                activeTab === 'dag' &&
                  'text-[var(--accent2)] border-b-[var(--accent2)] font-semibold'
              )}
              onClick={() => setActiveTab('dag')}
            >
              Element DAG
            </button>
          </div>
        </div>
        <div className="modal-body flex-1 overflow-y-auto p-5">
          {activeTab === 'general' ? (
            <div>
              <div className="grid grid-cols-2 gap-2.5">
                <div className="form-row mb-3.5">
                  <label className="block text-[11px] text-[var(--muted)] uppercase tracking-[0.5px] mb-1.5">
                    Key
                  </label>
                  <input
                    className="form-input w-full p-2 rounded-lg bg-[var(--surface)] border border-[var(--border)] text-[var(--text)] text-[13px] font-mono"
                    value={key}
                    onChange={e => setKey(e.target.value)}
                    disabled={!!existingTaxonomy}
                    placeholder="e.g. web-nav"
                  />
                </div>
                <div></div>
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
                    placeholder="Short tooltip"
                  />
                </div>
              </div>
              <div className="form-row mb-3.5">
                <label className="block text-[11px] text-[var(--muted)] uppercase tracking-[0.5px] mb-1.5">
                  Allowed Categories
                </label>
                <div className="flex flex-wrap gap-2 p-3 bg-[var(--surface)] border border-[var(--border)] rounded-xl min-h-[44px]">
                  {categories.map(cat => {
                    const isSelected = allowedCats.includes(cat.key);
                    return (
                      <button
                        key={cat.id}
                        className={cn(
                          'px-3 py-1 rounded-full text-[12px] border transition-all cursor-pointer',
                          isSelected
                            ? 'bg-[var(--accent2)] border-[var(--accent2)] text-white font-semibold shadow-sm'
                            : 'bg-transparent border-[var(--border)] text-[var(--muted)] hover:border-[var(--muted)]'
                        )}
                        onClick={() => {
                          if (isSelected) setAllowedCats(allowedCats.filter(c => c !== cat.key));
                          else setAllowedCats([...allowedCats, cat.key]);
                        }}
                      >
                        {cat.name}
                      </button>
                    );
                  })}
                </div>
              </div>
              <div className="form-row mb-3.5">
                <label className="block text-[11px] text-[var(--muted)] uppercase tracking-[0.5px] mb-1.5">
                  Description
                </label>
                <textarea
                  className="form-input w-full p-2 rounded-lg bg-[var(--surface)] border border-[var(--border)] text-[var(--text)] text-[13px] resize-vertical"
                  rows={2}
                  value={desc}
                  onChange={e => setDesc(e.target.value)}
                  placeholder="Longer description..."
                />
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
          ) : activeTab === 'properties' ? (
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
                        next[i].type = e.target.value as
                          | 'string'
                          | 'number'
                          | 'boolean'
                          | 'object'
                          | 'array';
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
          ) : (
            <div className="dag-tab-content h-full flex flex-col">
              <div className="mb-4 p-3 bg-[var(--surface2)] rounded-lg border border-[var(--border)]">
                <h4 className="text-[11px] text-[var(--muted)] uppercase tracking-[0.5px] mb-2">
                  Scope: Selected Categories
                </h4>
                <div className="flex flex-wrap gap-1.5">
                  {allowedCats.length > 0 ? (
                    allowedCats.map(catKey => (
                      <span
                        key={catKey}
                        className="px-2 py-0.5 bg-[var(--surface)] border border-[var(--border)] rounded-full text-[10px] text-[var(--accent2)] font-medium"
                      >
                        {categories.find(c => c.key === catKey)?.name || catKey}
                      </span>
                    ))
                  ) : (
                    <span className="text-[11px] text-[var(--muted)] italic">
                      No categories selected in General tab
                    </span>
                  )}
                </div>
              </div>
              <div className="flex-1 flex flex-col items-center justify-center border-2 border-dashed border-[var(--border)] rounded-xl bg-[var(--surface)] p-10 text-center">
                <div className="icon text-4xl mb-4 opacity-40">🌿</div>
                <h3 className="text-[15px] font-medium mb-2">DAG Relationship Editor</h3>
                <p className="text-[13px] text-[var(--muted)] max-w-[300px]">
                  Build Directed Acyclic Graphs by defining parent-child relationships between
                  elements of the selected categories.
                </p>
                <div className="mt-6 flex gap-3">
                  <button className="p-2 px-4 bg-[var(--surface2)] border border-[var(--border)] rounded-lg text-[12px] text-[var(--text)] cursor-not-allowed opacity-50">
                    Add Relationship
                  </button>
                  <button className="p-2 px-4 bg-[var(--surface2)] border border-[var(--border)] rounded-lg text-[12px] text-[var(--text)] cursor-not-allowed opacity-50">
                    Import Edges
                  </button>
                </div>
              </div>
            </div>
          )}
        </div>
        <div className="modal-footer p-3.5 px-6 border-t border-[var(--border)] flex items-center gap-2.5 shrink-0">
          {existingTaxonomy && onDelete && (
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
            {existingTaxonomy ? 'Save Changes' : 'Add Taxonomy'}
          </button>
        </div>
      </div>
    </div>
  );
};
