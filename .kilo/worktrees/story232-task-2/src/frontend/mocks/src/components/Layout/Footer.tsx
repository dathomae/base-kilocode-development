import React from 'react';

interface FooterProps {
  effectiveDate: string;
  onDateChange: (val: number) => void;
  helpMsg: string;
}

export const Footer: React.FC<FooterProps> = ({ effectiveDate, onDateChange, helpMsg }) => {
  return (
    <footer className="flex items-center h-[46px] bg-[var(--surface)] shrink-0 border-t border-[var(--border)] text-[12px]">
      <div className="w-[240px] shrink-0 px-3.5 flex items-center gap-2 border-r border-[var(--border)]">
        <label className="text-[10px] text-[var(--muted)] whitespace-nowrap">Effective</label>
        <input
          type="range"
          min="0"
          max="100"
          defaultValue="100"
          className="flex-1 accent-[var(--accent)] cursor-pointer"
          onChange={e => onDateChange(parseInt(e.target.value))}
        />
        <span className="text-[10px] text-[var(--accent2)] whitespace-nowrap min-w-[80px]">
          {effectiveDate}
        </span>
      </div>
      <div className="flex-1 flex items-center gap-2.5 px-4 overflow-hidden">
        <span className="text-[var(--muted)] whitespace-nowrap">© 2026 Timecone.ai</span>
        <span className="flex-1 text-[var(--muted)] italic overflow-hidden text-ellipsis whitespace-nowrap">
          {helpMsg}
        </span>
        <span className="text-[var(--accent2)] whitespace-nowrap hover:underline cursor-pointer">
          Help ›
        </span>
      </div>
    </footer>
  );
};
