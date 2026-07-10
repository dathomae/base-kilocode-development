import React from 'react';

export const Header: React.FC = () => {
  return (
    <header className="flex items-center justify-between bg-[var(--surface)] px-5 h-[54px] shrink-0 border-b border-[var(--border)] shadow-[0_2px_12px_rgba(0,0,0,0.5)]">
      <div className="flex items-center gap-3.5">
        <img
          src="/timecone-on-background.png"
          alt="Timecone Branding"
          className="w-[34px] h-[34px] rounded-lg object-cover shadow-sm"
          referrerPolicy="no-referrer"
        />
        <div className="text-[17px] font-semibold">
          Timecone<span className="text-[var(--accent2)]">.ai</span>
        </div>
      </div>
      <div className="flex items-center gap-3.5">
        <span className="text-sm text-[var(--muted)]">👋 Alex Johnson</span>
        <div
          title="EN-US"
          className="flex items-center justify-center w-7 h-7 bg-[var(--surface2)] border border-[var(--border)] rounded-full text-[14px] cursor-help transition-all hover:border-[var(--accent2)]"
        >
          🇺🇸
        </div>
        <div className="relative">
          <button className="bg-none border-none cursor-pointer text-[var(--muted)] text-xl transition-colors hover:text-[var(--accent2)]">
            ⚙
          </button>
        </div>
      </div>
    </header>
  );
};
