/**
 * Shared Tabulator column helpers and DOM utilities used across table components.
 *
 * Cell display (from screenshot spec):
 *   [🌐 N]  Display Name
 *
 * Hover tooltip (appears below the cell, position:fixed so it escapes Tabulator
 * overflow clipping):
 *   ┌──────────────────────────────┐
 *   │ [key-badge]        More...   │
 *   │ ────────────────────────────  │
 *   │ tooltip text                  │
 *   └──────────────────────────────┘
 *
 * Clicking "More..." fires the onMore callback with (id, name, fullDesc) and
 * is handled by the component via attachMoreLinkListener.
 */

type DescCallback = (id: string, name: string, desc: string) => void;

// ---------------------------------------------------------------------------
// getCommonNameColumn
// ---------------------------------------------------------------------------

/**
 * Builds a standard name column config for Tabulator.
 *
 * @param title  - Column header text
 * @param onMore - Called when the user clicks the "More…" link (unused here –
 *                 the actual invocation is in attachMoreLinkListener)
 * @param field  - Row data field used as the display name (default: 'name')
 */
export function getCommonNameColumn(title: string, _onMore: DescCallback, field: string = 'name') {
  return {
    title,
    field,
    widthGrow: 2,
    formatter: (cell: { getRow: () => { getData: () => Record<string, unknown> } }) => {
      const data = cell.getRow().getData();
      const id = String(data.id ?? '');
      const key = String(data.key ?? '');
      let name = String(data[field] ?? '');
      let tip = String(data.tooltip ?? '');
      let desc = String(data.desc ?? '');
      const locs = (data._locales as unknown[]) ?? [];

      const targetLocale = 'US-EN';
      const [targetLang, targetRegion] = targetLocale.split('-');

      let matchedLoc = null;
      if (locs.length > 0) {
        matchedLoc =
          locs.find(
            l =>
              (l as Record<string, string>).lang?.toLowerCase() === targetLang.toLowerCase() &&
              (l as Record<string, string>).region?.toLowerCase() === targetRegion.toLowerCase()
          ) ||
          locs.find(
            l => (l as Record<string, string>).lang?.toLowerCase() === targetLang.toLowerCase()
          ) ||
          locs[0];
      }

      if (matchedLoc) {
        name = (matchedLoc as Record<string, string>).name
          ? String((matchedLoc as Record<string, string>).name)
          : name;
        tip = (matchedLoc as Record<string, string>).tooltip
          ? String((matchedLoc as Record<string, string>).tooltip)
          : tip;
        desc = (matchedLoc as Record<string, string>).desc
          ? String((matchedLoc as Record<string, string>).desc)
          : desc;
      }

      // ── locale badge (teal) ──────────────────────────────────────────
      const locBadge =
        locs.length > 0
          ? `<button
             class="badge-btn badge-loc"
             data-id="${id}"
             data-rowid="${id}"
             data-taxid="${id}"
           >🌐 ${locs.length}</button>`
          : '';

      // ── hover-tooltip card (rendered hidden; shown by attachMoreLinkListener)
      // The card uses position:fixed so it escapes Tabulator's overflow:hidden.
      // The top/left are set dynamically via JS in attachMoreLinkListener.
      const keyBadge = key
        ? `<span class="badge-btn badge-prop" style="pointer-events:none;">${key}</span>`
        : '';

      const moreLink = desc
        ? `<a
             class="more-link"
             data-id="${id}"
             data-name="${encodeURIComponent(name)}"
             data-desc="${encodeURIComponent(desc)}"
             style="color:var(--accent);font-size:12px;text-decoration:underline;cursor:pointer;margin-left:auto;"
           >More...</a>`
        : '';

      const tipCard =
        key || tip || desc
          ? `<div
             class="name-tip"
             style="display: none; position: fixed; z-index: 9999; min-width: 200px; max-width: 300px; background: var(--surface2); border: 1px solid var(--border); border-radius: 8px; padding: 10px 12px; box-shadow: 0 4px 20px rgba(0,0,0,0.55); pointer-events: auto;"
           >
             <div style="display:flex;align-items:center;gap:8px;margin-bottom:6px;">
               ${keyBadge}
               ${moreLink}
             </div>
             <hr style="border:none;border-top:1px solid var(--border);margin:0 0 8px;" />
             <div style="font-size:12px;color:var(--text);line-height:1.5;">${tip}</div>
           </div>`
          : '';

      // ── cell content ──────────────────────────────────────────────────
      return `
        <div class="name-col-wrap" style="display:inline-flex;vertical-align:middle;align-items:center;gap:6px;position:relative;">
          ${locBadge}
          <span style="font-weight:500;">${name}</span>
          ${tipCard}
        </div>
      `;
    },
  };
}

// ---------------------------------------------------------------------------
// attachMoreLinkListener
// ---------------------------------------------------------------------------

export function attachMoreLinkListener(tableEl: HTMLElement, onMore: DescCallback): () => void {
  let activeTip: HTMLElement | null = null;
  let activeAncestor: HTMLElement | null = null;
  let hideTimer: ReturnType<typeof setTimeout> | null = null;

  function clearHideTimer() {
    if (hideTimer !== null) {
      clearTimeout(hideTimer);
      hideTimer = null;
    }
  }

  function showTip(tip: HTMLElement, anchor: HTMLElement) {
    clearHideTimer();

    // If we're showing a new tip but haven't hidden the old one yet, hide it now
    if (activeTip && activeTip !== tip) {
      hideTip();
    }

    // Move to body to escape any CSS transforms or overflow clipping
    if (tip.parentElement !== document.body) {
      activeAncestor = anchor;
      document.body.appendChild(tip);
    }
    const rect = anchor.getBoundingClientRect();
    tip.style.top = rect.bottom + 6 + 'px';
    tip.style.left = rect.left + 'px';
    tip.style.display = 'block';
    activeTip = tip;
  }

  function hideTip() {
    if (activeTip) {
      activeTip.style.display = 'none';
      if (activeAncestor && activeTip.parentElement === document.body) {
        // Move back to anchor so it's found by querySelector on next hover
        activeAncestor.appendChild(activeTip);
      }
      activeTip = null;
      activeAncestor = null;
    }
  }

  function scheduleHide() {
    clearHideTimer();
    hideTimer = setTimeout(hideTip, 120);
  }

  // mouseover – show tooltip when entering .name-col-wrap
  const onOver = (e: MouseEvent) => {
    const wrap = (e.target as HTMLElement).closest('.name-col-wrap') as HTMLElement | null;
    if (!wrap) return;
    const tip = wrap.querySelector('.name-tip') as HTMLElement | null;
    if (tip) showTip(tip, wrap);
  };

  // mouseout – hide tooltip when leaving the wrap or the tip itself
  const onOut = (e: MouseEvent) => {
    const wrap = (e.target as HTMLElement).closest('.name-col-wrap');
    if (!wrap) return;
    const related = e.relatedTarget as HTMLElement | null;
    // Stay visible if moving to the tooltip card
    if (related && activeTip && activeTip.contains(related)) {
      clearHideTimer();
      return;
    }
    scheduleHide();
  };

  // mouseover on the tip card itself – cancel pending hide
  const onTipOver = (e: MouseEvent) => {
    const tip = (e.target as HTMLElement).closest('.name-tip');
    if (tip && activeTip && tip === activeTip) {
      clearHideTimer();
    }
  };

  // mouseout on the tip card – schedule hide
  const onTipOut = (e: MouseEvent) => {
    const target = e.target as HTMLElement;
    const tip = target.closest('.name-tip');
    if (!tip || tip !== activeTip) return;
    const related = e.relatedTarget as HTMLElement | null;
    // Stay visible if moving back to the wrap
    if (related && activeAncestor && activeAncestor.contains(related)) {
      clearHideTimer();
      return;
    }
    scheduleHide();
  };

  // click – fire the More callback
  const onClick = (e: MouseEvent) => {
    const link = (e.target as HTMLElement).closest('.more-link') as HTMLElement | null;
    if (!link) return;
    // ensure this click happened within our mapped tooltip
    if (!activeTip || !activeTip.contains(link)) return;

    hideTip();
    clearHideTimer();
    onMore(
      link.dataset.id ?? '',
      decodeURIComponent(link.dataset.name ?? ''),
      decodeURIComponent(link.dataset.desc ?? '')
    );
  };

  tableEl.addEventListener('mouseover', onOver);
  tableEl.addEventListener('mouseout', onOut);

  // tooltip events bind to document as tip is moved to body
  document.addEventListener('mouseover', onTipOver);
  document.addEventListener('mouseout', onTipOut);
  document.addEventListener('click', onClick);

  return () => {
    tableEl.removeEventListener('mouseover', onOver);
    tableEl.removeEventListener('mouseout', onOut);
    document.removeEventListener('mouseover', onTipOver);
    document.removeEventListener('mouseout', onTipOut);
    document.removeEventListener('click', onClick);
    clearHideTimer();
    hideTip();
  };
}
