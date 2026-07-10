import { X } from 'lucide-react';
import React, { useEffect, useState } from 'react';
import Markdown from 'react-markdown';
import { cn } from '../../lib/utils';
import { Category, Converter, Measure } from '../../types';

interface ConverterModalProps {
  isOpen: boolean;
  onClose: () => void;
  onConfirm: (data: Converter) => void;
  onDelete?: (id: string) => void;
  existingConverter?: Converter | null;
  measures: Measure[];
  categories: Category[];
}

export const ConverterModal: React.FC<ConverterModalProps> = ({
  isOpen,
  onClose,
  onConfirm,
  onDelete,
  existingConverter,
  measures,
  categories,
}) => {
  const [type, setType] = useState<'Measure' | 'Unit'>('Measure');
  const [assignment, setAssignment] = useState({ id: '', name: '', alias: '' });
  const [input, setInput] = useState({ id: '', name: '', alias: '' });
  const [calculation, setCalculation] = useState('');

  useEffect(() => {
    setTimeout(() => {
      if (existingConverter) {
        setType(existingConverter.type);
        setAssignment(existingConverter.assignment);
        setInput(existingConverter.input);
        setCalculation(existingConverter.calculation);
      } else {
        setType('Measure');
        setAssignment({ id: '', name: '', alias: '' });
        setInput({ id: '', name: '', alias: '' });
        setCalculation('');
      }
    }, 0);
  }, [existingConverter, isOpen]);

  if (!isOpen) return null;

  const unitCategories = categories.filter(
    c => c.isBaseUnit || c.key === 'channel' || c.key === 'product_class'
  );
  const allUnits = unitCategories.flatMap(c => c._children);

  const proposeAlias = (name: string, currentAliases: string[]) => {
    if (!name) return '';
    const base = name.charAt(0).toLowerCase();
    let alias = base;
    let counter = 1;
    while (currentAliases.includes(alias)) {
      alias = base + counter;
      counter++;
    }
    return alias;
  };

  const handleSelection = (field: 'assignment' | 'input', id: string) => {
    let name = '';
    if (type === 'Measure') {
      name = measures.find(m => m.id === id)?.name || '';
    } else {
      name = allUnits.find(u => u.id === id)?.name || '';
    }

    const currentAliases = [field === 'assignment' ? input.alias : assignment.alias].filter(
      Boolean
    );

    const alias = proposeAlias(name, currentAliases);

    if (field === 'assignment') {
      setAssignment({ id, name, alias });
    } else {
      setInput({ id, name, alias });
    }
  };

  const handleConfirm = () => {
    onConfirm({
      id: existingConverter?.id || `conv-${Date.now()}`,
      type,
      assignment,
      input,
      calculation,
    });
  };

  const isValid = assignment.id && input.id && calculation.trim();

  return (
    <div className="fixed inset-0 z-[1000] flex items-center justify-center bg-black/50 backdrop-blur-sm animate-in fade-in duration-200">
      <div className="bg-[var(--surface)] border border-[var(--border)] rounded-xl shadow-2xl w-full max-w-2xl overflow-hidden flex flex-col max-h-[90vh]">
        <div className="p-5 border-b border-[var(--border)] flex items-center justify-between bg-[var(--surface2)]/50">
          <h2 className="text-xl font-semibold tracking-tight">
            {existingConverter ? 'Edit Converter' : 'Add Converter'}
          </h2>
          <button
            onClick={onClose}
            className="p-1.5 rounded-lg hover:bg-[var(--surface2)] text-[var(--muted)] transition-colors"
          >
            <X size={20} />
          </button>
        </div>

        <div className="flex-1 overflow-y-auto p-8 space-y-8">
          {/* Type Toggle */}
          <div className="space-y-3">
            <label className="text-xs font-bold uppercase tracking-wider text-[var(--muted)]">
              Converter Type
            </label>
            <div className="flex p-1 bg-[var(--surface2)] rounded-xl border border-[var(--border)] w-fit">
              {(['Measure', 'Unit'] as const).map(t => (
                <button
                  key={t}
                  onClick={() => {
                    setType(t);
                    setAssignment({ id: '', name: '', alias: '' });
                    setInput({ id: '', name: '', alias: '' });
                  }}
                  className={cn(
                    'px-6 py-2 rounded-lg text-sm font-medium transition-all',
                    type === t
                      ? 'bg-[var(--accent)] text-white shadow-lg'
                      : 'text-[var(--muted)] hover:text-[var(--text)]'
                  )}
                >
                  {t}
                </button>
              ))}
            </div>
          </div>

          <div className="grid grid-cols-2 gap-8">
            {/* Assignment */}
            <div className="space-y-4">
              <div className="space-y-2">
                <label className="text-xs font-bold uppercase tracking-wider text-[var(--muted)]">
                  Assignment ({type})
                </label>
                <select
                  value={assignment.id}
                  onChange={e => handleSelection('assignment', e.target.value)}
                  className="w-full bg-[var(--surface2)] border border-[var(--border)] rounded-lg px-4 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-[var(--accent)]/50 transition-all"
                >
                  <option value="">Select {type}</option>
                  {type === 'Measure'
                    ? measures.map(m => (
                        <option key={m.id} value={m.id}>
                          {m.name}
                        </option>
                      ))
                    : allUnits.map(u => (
                        <option key={u.id} value={u.id}>
                          {u.name}
                        </option>
                      ))}
                </select>
              </div>
              {assignment.id && (
                <div className="space-y-2 animate-in slide-in-from-top-2 duration-200">
                  <label className="text-xs font-bold uppercase tracking-wider text-[var(--muted)]">
                    Alias
                  </label>
                  <input
                    value={assignment.alias}
                    onChange={e => setAssignment(prev => ({ ...prev, alias: e.target.value }))}
                    className="w-full bg-[var(--surface2)] border border-[var(--border)] rounded-lg px-4 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-[var(--accent)]/50 transition-all font-mono text-[var(--warn)]"
                    placeholder="Alias"
                  />
                </div>
              )}
            </div>

            {/* Input */}
            <div className="space-y-4">
              <div className="space-y-2">
                <label className="text-xs font-bold uppercase tracking-wider text-[var(--muted)]">
                  Input ({type})
                </label>
                <select
                  value={input.id}
                  onChange={e => handleSelection('input', e.target.value)}
                  className="w-full bg-[var(--surface2)] border border-[var(--border)] rounded-lg px-4 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-[var(--accent)]/50 transition-all"
                >
                  <option value="">Select {type}</option>
                  {type === 'Measure'
                    ? measures.map(m => (
                        <option key={m.id} value={m.id}>
                          {m.name}
                        </option>
                      ))
                    : allUnits.map(u => (
                        <option key={u.id} value={u.id}>
                          {u.name}
                        </option>
                      ))}
                </select>
              </div>
              {input.id && (
                <div className="space-y-2 animate-in slide-in-from-top-2 duration-200">
                  <label className="text-xs font-bold uppercase tracking-wider text-[var(--muted)]">
                    Alias
                  </label>
                  <input
                    value={input.alias}
                    onChange={e => setInput(prev => ({ ...prev, alias: e.target.value }))}
                    className="w-full bg-[var(--surface2)] border border-[var(--border)] rounded-lg px-4 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-[var(--accent)]/50 transition-all font-mono text-[var(--warn)]"
                    placeholder="Alias"
                  />
                </div>
              )}
            </div>
          </div>

          {/* Calculation */}
          <div className="space-y-4">
            <div className="space-y-2">
              <label className="text-xs font-bold uppercase tracking-wider text-[var(--muted)]">
                Calculation
              </label>
              <textarea
                value={calculation}
                onChange={e => setCalculation(e.target.value)}
                rows={6}
                className="w-full bg-[var(--surface2)] border border-[var(--border)] rounded-lg px-4 py-3 text-sm focus:outline-none focus:ring-2 focus:ring-[var(--accent)]/50 transition-all font-mono resize-none leading-relaxed"
                placeholder={`e.g. if (\`${assignment.alias || 'a'}\` > 0) {\n  return \`${input.alias || 'i'}\` * 1.2;\n}`}
              />
              <p className="text-[10px] text-[var(--muted)]">
                Enter the conversion logic. Wrap aliases in backticks like `\`$
                {assignment.alias || 'a'}\``.{' '}
              </p>
            </div>

            {calculation && (
              <div className="space-y-2 animate-in fade-in slide-in-from-top-2 duration-300">
                <label className="text-xs font-bold uppercase tracking-wider text-[var(--muted)]">
                  Preview
                </label>
                <div className="p-4 bg-[var(--surface2)] border border-[var(--border)] rounded-lg markdown-body overflow-x-auto">
                  <Markdown
                    components={{
                      code({ children }) {
                        const content = String(children).trim();
                        const isAlias = content === assignment.alias || content === input.alias;
                        return (
                          <code
                            className={cn(
                              'px-1.5 py-0.5 rounded text-[13px] font-mono',
                              isAlias
                                ? 'bg-[var(--warn)]/10 text-[var(--warn)] border border-[var(--warn)]/20'
                                : 'bg-[var(--surface)] border border-[var(--border)] text-[var(--accent2)]'
                            )}
                          >
                            {children}
                          </code>
                        );
                      },
                    }}
                  >
                    {calculation}
                  </Markdown>
                </div>
              </div>
            )}
          </div>
        </div>

        <div className="p-6 border-t border-[var(--border)] bg-[var(--surface2)]/50 flex items-center justify-between">
          <div>
            {existingConverter && onDelete && (
              <button
                onClick={() => onDelete(existingConverter.id)}
                className="px-5 py-2 rounded-lg border border-red-500/20 text-red-500 hover:bg-red-500/10 transition-all text-sm font-medium"
              >
                Delete Converter
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
              {existingConverter ? 'Save Changes' : 'Add Converter'}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};
