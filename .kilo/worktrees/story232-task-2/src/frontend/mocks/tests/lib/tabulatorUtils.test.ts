/**
 * Regression tests for tabulatorUtils.ts
 *
 * These pin the exact cell HTML shape and behaviour so a visual agent cannot
 * silently break the display by regenerating the file.
 *
 * SPEC (from screenshot):
 *  ┌─ cell ────────────────────────────────────────────────┐
 *  │  [🌐 N]  Display Name                                 │
 *  └───────────────────────────────────────────────────────┘
 *
 *  hover tooltip card (position:fixed, shown via JS):
 *  ┌──────────────────────────────────────────┐
 *  │  [key-badge]               More...       │
 *  │  ────────────────────────────────────── │
 *  │  tooltip text                            │
 *  └──────────────────────────────────────────┘
 */

import { beforeEach, describe, expect, it, vi } from 'vitest';
import { attachMoreLinkListener, getCommonNameColumn } from '../../src/lib/tabulatorUtils';

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

/** Parse the HTML string a formatter returns into a live DOM element */
function renderFormatter(html: string): HTMLElement {
  const div = document.createElement('div');
  div.innerHTML = html.trim();
  return div.firstElementChild as HTMLElement;
}

/** Minimal cell mock that Tabulator formatters receive */
function makeCell(data: Record<string, unknown>): {
  getRow: () => { getData: () => Record<string, unknown> };
} {
  return {
    getRow: () => ({ getData: () => data }),
  };
}

function format(field: string, data: Record<string, unknown>): HTMLElement {
  const col = getCommonNameColumn('Name', vi.fn(), field);
  const html = (col as { formatter: (c: unknown) => string }).formatter(makeCell(data));

  return renderFormatter(html);
}

// ---------------------------------------------------------------------------
// Fixture data
// ---------------------------------------------------------------------------

const PLAN_ROW = {
  id: 1001,
  key: '5w_rev',
  label: '5-week Review',
  tooltip: 'Short-term performance review',
  desc: 'A detailed look at the last 5 weeks of performance.',
  _locales: [
    {
      lang: 'en',
      region: '',
      name: '5-week Review',
      tooltip: '',
      desc: 'A detailed look at the last 5 weeks of performance.',
    },
  ],
};

const CAT_ROW = {
  id: 'cat-channel',
  key: 'channel',
  name: 'Channel',
  tooltip: 'Sales channel',
  desc: 'The route through which product reaches the customer.',
  _locales: [
    { lang: 'en', region: '', name: 'Channel', tooltip: 'Sales channel', desc: '' },
    { lang: 'fr', region: 'FR', name: 'Canal', tooltip: 'Canal de vente', desc: '' },
  ],
};

const EMPTY_ROW = {
  id: 'x',
  key: '',
  name: 'X',
  tooltip: '',
  desc: '',
  _locales: [],
};

// ---------------------------------------------------------------------------
// Cell display structure
// ---------------------------------------------------------------------------

describe('getCommonNameColumn – cell display', () => {
  it('root element has class name-col-wrap', () => {
    const el = format('label', PLAN_ROW);
    expect(el.classList.contains('name-col-wrap')).toBe(true);
  });

  it('renders the display name as visible text', () => {
    const el = format('label', PLAN_ROW);
    expect(el.textContent).toContain('5-week Review');
  });

  it('renders a locale badge when _locales is non-empty', () => {
    const el = format('label', PLAN_ROW);
    const badge = el.querySelector('.badge-loc');
    expect(badge).not.toBeNull();
  });

  it('locale badge shows 🌐 and the locale count', () => {
    const el = format('name', CAT_ROW);
    const badge = el.querySelector('.badge-loc')!;
    expect(badge.textContent).toContain('🌐');
    expect(badge.textContent).toContain('2');
  });

  it('does NOT render a locale badge when _locales is empty', () => {
    const el = format('name', EMPTY_ROW);
    expect(el.querySelector('.badge-loc')).toBeNull();
  });

  it('locale badge carries data-id, data-rowid, and data-taxid for component compatibility', () => {
    const el = format('name', CAT_ROW);
    const badge = el.querySelector('.badge-loc') as HTMLElement;
    expect(badge.dataset.id).toBe('cat-channel');
    expect(badge.dataset.rowid).toBe('cat-channel');
    expect(badge.dataset.taxid).toBe('cat-channel');
  });

  it('uses the specified field (label vs name)', () => {
    const elLabel = format('label', PLAN_ROW);
    expect(elLabel.textContent).toContain('5-week Review');

    const catEl = format('name', CAT_ROW);
    expect(catEl.textContent).toContain('Channel');
  });
});

// ---------------------------------------------------------------------------
// Tooltip card structure
// ---------------------------------------------------------------------------

describe('getCommonNameColumn – tooltip card', () => {
  it('renders a hidden .name-tip element inside the cell', () => {
    const el = format('label', PLAN_ROW);
    const tip = el.querySelector('.name-tip') as HTMLElement | null;
    expect(tip).not.toBeNull();
    expect(tip!.getAttribute('style')).toContain('display: none');
  });

  it('tooltip card contains the key in an orange .badge-prop badge', () => {
    const el = format('label', PLAN_ROW);
    const tip = el.querySelector('.name-tip')!;
    const keyBadge = tip.querySelector('.badge-prop');
    expect(keyBadge).not.toBeNull();
    expect(keyBadge!.textContent).toContain('5w_rev');
  });

  it('tooltip card contains a .more-link anchor', () => {
    const el = format('label', PLAN_ROW);
    const moreLink = el.querySelector('.name-tip .more-link') as HTMLElement | null;
    expect(moreLink).not.toBeNull();
  });

  it('more-link carries encoded name and desc data-attributes', () => {
    const el = format('label', PLAN_ROW);
    const link = el.querySelector('.more-link') as HTMLElement;
    expect(decodeURIComponent(link.dataset.name ?? '')).toBe('5-week Review');
    expect(decodeURIComponent(link.dataset.desc ?? '')).toContain('last 5 weeks');
  });

  it('tooltip card contains a horizontal rule separating key and tooltip text', () => {
    const el = format('label', PLAN_ROW);
    const tip = el.querySelector('.name-tip')!;
    expect(tip.querySelector('hr')).not.toBeNull();
  });

  it('tooltip card shows the tooltip text below the hr', () => {
    const el = format('label', PLAN_ROW);
    const tip = el.querySelector('.name-tip')!;
    expect(tip.textContent).toContain('Short-term performance review');
  });

  it('tooltip uses position:fixed so it escapes Tabulator overflow clipping', () => {
    const el = format('label', PLAN_ROW);
    const tip = el.querySelector('.name-tip') as HTMLElement;
    expect(tip.getAttribute('style')).toContain('position: fixed');
  });

  it('NO tooltip card is rendered for rows with no key/tooltip/desc', () => {
    const el = format('name', EMPTY_ROW);
    expect(el.querySelector('.name-tip')).toBeNull();
  });
});

// ---------------------------------------------------------------------------
// attachMoreLinkListener – More link callback
// ---------------------------------------------------------------------------

describe('attachMoreLinkListener – More link click', () => {
  let tableEl: HTMLElement;
  let onMore: ReturnType<typeof vi.fn>;
  let cleanup: () => void;

  beforeEach(() => {
    tableEl = document.createElement('div');
    document.body.appendChild(tableEl);
    onMore = vi.fn();
    cleanup = attachMoreLinkListener(tableEl, onMore);
  });

  it('calls onMore with decoded id, name, desc when a .more-link is clicked', () => {
    const wrap = document.createElement('div');
    wrap.className = 'name-col-wrap';
    const tip = document.createElement('div');
    tip.className = 'name-tip';
    tip.style.display = 'none';

    const link = document.createElement('a');
    link.className = 'more-link';
    link.dataset.id = '1001';
    link.dataset.name = encodeURIComponent('5-week Review');
    link.dataset.desc = encodeURIComponent('A detailed look.');

    tip.appendChild(link);
    wrap.appendChild(tip);
    tableEl.appendChild(wrap);

    wrap.dispatchEvent(new MouseEvent('mouseover', { bubbles: true }));
    link.dispatchEvent(new MouseEvent('click', { bubbles: true }));

    expect(onMore).toHaveBeenCalledWith('1001', '5-week Review', 'A detailed look.');
  });

  it('does not fire onMore for clicks outside .more-link', () => {
    const span = document.createElement('span');
    tableEl.appendChild(span);
    span.dispatchEvent(new MouseEvent('click', { bubbles: true }));
    expect(onMore).not.toHaveBeenCalled();
  });

  it('cleanup removes the listener so subsequent clicks are ignored', () => {
    const wrap = document.createElement('div');
    wrap.className = 'name-col-wrap';
    const tip = document.createElement('div');
    tip.className = 'name-tip';
    tip.style.display = 'none';

    const link = document.createElement('a');
    link.className = 'more-link';
    link.dataset.id = '99';
    link.dataset.name = encodeURIComponent('X');
    link.dataset.desc = encodeURIComponent('Y');

    tip.appendChild(link);
    wrap.appendChild(tip);
    tableEl.appendChild(wrap);

    wrap.dispatchEvent(new MouseEvent('mouseover', { bubbles: true }));

    cleanup();
    link.dispatchEvent(new MouseEvent('click', { bubbles: true }));
    expect(onMore).not.toHaveBeenCalled();
  });
});

// ---------------------------------------------------------------------------
// attachMoreLinkListener – tooltip show/hide on hover
// ---------------------------------------------------------------------------

describe('attachMoreLinkListener – tooltip positioning', () => {
  let tableEl: HTMLElement;
  let cleanup: () => void;
  let wrap: HTMLElement;
  let tip: HTMLElement;

  beforeEach(() => {
    tableEl = document.createElement('div');
    document.body.appendChild(tableEl);

    wrap = document.createElement('div');
    wrap.className = 'name-col-wrap';
    tip = document.createElement('div');
    tip.className = 'name-tip';
    tip.style.display = 'none';
    wrap.appendChild(tip);
    tableEl.appendChild(wrap);

    cleanup = attachMoreLinkListener(tableEl, vi.fn());
  });

  it('shows the .name-tip on mouseover of .name-col-wrap', () => {
    wrap.dispatchEvent(new MouseEvent('mouseover', { bubbles: true }));
    expect(tip.style.display).toBe('block');
  });

  it('hides the .name-tip after mouseout of .name-col-wrap (after delay)', async () => {
    wrap.dispatchEvent(new MouseEvent('mouseover', { bubbles: true }));
    expect(tip.style.display).toBe('block');

    wrap.dispatchEvent(new MouseEvent('mouseout', { bubbles: true, relatedTarget: document.body }));
    // hide is scheduled with a small timeout; wait for it
    await new Promise(r => setTimeout(r, 200));
    expect(tip.style.display).toBe('none');
  });

  it('cleanup stops future hover events from showing the tooltip', () => {
    cleanup();
    wrap.dispatchEvent(new MouseEvent('mouseover', { bubbles: true }));
    expect(tip.style.display).toBe('none');
  });
});
