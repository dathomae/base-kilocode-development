import '@testing-library/jest-dom';
import { vi } from 'vitest';

/**
 * Tabulator is a DOM-heavy library that tries to measure element dimensions on
 * initialisation.  Stub it out entirely so tests that import components
 * containing Tabulator don't crash in happy-dom.
 */
vi.mock('tabulator-tables', () => {
  class TabulatorFull {
    on() {}
    destroy() {}
    setData() {}
    setFilter() {}
    getSelectedRows() {
      return [];
    }
    getRow() {
      return null;
    }
    selectRow() {}
    deselectRow() {}
    redraw() {}
  }
  return { TabulatorFull };
});

/**
 * @xyflow/react uses ResizeObserver which isn't available in happy-dom.
 */
if (typeof globalThis.ResizeObserver === 'undefined') {
  globalThis.ResizeObserver = class {
    observe() {}
    unobserve() {}
    disconnect() {}
  };
}
