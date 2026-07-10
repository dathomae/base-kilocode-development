/**
 * DEFAULT_METRICS logic tests
 *
 * DEFAULT_METRICS(planId, label) is a pure function defined in App.tsx.
 * It selects which initial metrics a plan gets based on the plan label text.
 * We mirror the exact function here so it can be tested in isolation.
 */

import { describe, it, expect } from 'vitest';
import type { Metric } from '../../src/types';

// ---------------------------------------------------------------------------
// Mirror of DEFAULT_METRICS from App.tsx (must stay in sync)
// ---------------------------------------------------------------------------

const DEFAULT_METRICS = (planId: number, label: string): Metric[] => {
  const seed = planId * 1000;
  if (label.toLowerCase().includes('week') && label.toLowerCase().includes('review')) {
    return [
      {
        id: seed + 1,
        measure: 'inventory',
        dims: ['country', 'region', 'warehouse'],
        grain: 'daily',
      },
      { id: seed + 2, measure: 'demand', dims: ['category', 'country', 'region'], grain: 'weekly' },
    ];
  }
  if (label.toLowerCase().includes('year')) {
    return [
      { id: seed + 1, measure: 'revenue', dims: ['country', 'channel'], grain: 'monthly' },
      { id: seed + 2, measure: 'cost', dims: ['category', 'country'], grain: 'monthly' },
      {
        id: seed + 3,
        measure: 'demand',
        dims: ['category', 'country', 'product'],
        grain: 'weekly',
      },
    ];
  }
  return [{ id: seed + 1, measure: 'inventory', dims: ['country'], grain: 'daily' }];
};

// ---------------------------------------------------------------------------
// "week + review" label branch
// ---------------------------------------------------------------------------

describe('DEFAULT_METRICS – week-review branch', () => {
  it('"5-week Review" returns 2 metrics', () => {
    const result = DEFAULT_METRICS(1001, '5-week Review');
    expect(result).toHaveLength(2);
  });

  it('"5-week Review" first metric is inventory / daily', () => {
    const result = DEFAULT_METRICS(1001, '5-week Review');
    expect(result[0].measure).toBe('inventory');
    expect(result[0].grain).toBe('daily');
  });

  it('"5-week Review" second metric is demand / weekly', () => {
    const result = DEFAULT_METRICS(1001, '5-week Review');
    expect(result[1].measure).toBe('demand');
    expect(result[1].grain).toBe('weekly');
  });

  it('"13-week Review" also triggers the week-review branch', () => {
    const result = DEFAULT_METRICS(1002, '13-week Review');
    expect(result).toHaveLength(2);
  });

  it('"5-week Forecast" does NOT trigger the week-review branch (missing "review")', () => {
    const result = DEFAULT_METRICS(1005, '5-week Forecast');
    expect(result).toHaveLength(1); // falls through to default
  });
});

// ---------------------------------------------------------------------------
// "year" label branch
// ---------------------------------------------------------------------------

describe('DEFAULT_METRICS – year branch', () => {
  it('"1-year Review" returns 3 metrics', () => {
    const result = DEFAULT_METRICS(1003, '1-year Review');
    expect(result).toHaveLength(3);
  });

  it('"1-year Review" includes revenue, cost, and demand measures', () => {
    const result = DEFAULT_METRICS(1003, '1-year Review');
    const measures = result.map(m => m.measure);
    expect(measures).toContain('revenue');
    expect(measures).toContain('cost');
    expect(measures).toContain('demand');
  });

  it('"1-year Review" metrics all have monthly or weekly grain', () => {
    const result = DEFAULT_METRICS(1003, '1-year Review');
    result.forEach(m => {
      expect(['monthly', 'weekly']).toContain(m.grain);
    });
  });

  it('"1-year Forecast" also triggers the year branch', () => {
    const result = DEFAULT_METRICS(1006, '1-year Forecast');
    expect(result).toHaveLength(3);
  });

  it('"3-year Review" triggers the year branch', () => {
    const result = DEFAULT_METRICS(1004, '3-year Review');
    expect(result).toHaveLength(3);
  });
});

// ---------------------------------------------------------------------------
// Default (single metric) branch
// ---------------------------------------------------------------------------

describe('DEFAULT_METRICS – default branch', () => {
  it('unrecognised label returns 1 metric', () => {
    const result = DEFAULT_METRICS(9, 'Monthly Report');
    expect(result).toHaveLength(1);
  });

  it('default metric is inventory / daily / country', () => {
    const result = DEFAULT_METRICS(9, 'Monthly Report');
    expect(result[0].measure).toBe('inventory');
    expect(result[0].grain).toBe('daily');
    expect(result[0].dims).toEqual(['country']);
  });
});

// ---------------------------------------------------------------------------
// Seed / ID calculation
// ---------------------------------------------------------------------------

describe('DEFAULT_METRICS – ID seed', () => {
  it('metric IDs are based on planId * 1000 + position', () => {
    const r1 = DEFAULT_METRICS(2, '5-week Review');
    expect(r1[0].id).toBe(2001);
    expect(r1[1].id).toBe(2002);
  });

  it('different planIds produce non-overlapping metric ID ranges', () => {
    const r1 = DEFAULT_METRICS(1, '5-week Review');
    const r2 = DEFAULT_METRICS(2, '5-week Review');
    const ids1 = new Set(r1.map(m => m.id));
    const ids2 = new Set(r2.map(m => m.id));
    const overlap = [...ids1].filter(id => ids2.has(id));
    expect(overlap).toHaveLength(0);
  });
});
