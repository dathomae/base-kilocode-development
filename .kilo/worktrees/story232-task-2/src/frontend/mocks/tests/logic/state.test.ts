/**
 * App state-handler tests
 *
 * The handlers in App.tsx are closures around React state.  We extract the
 * pure transformation logic as standalone functions here so they can be
 * exercised in isolation, without rendering the component.
 *
 * Each function mirrors the exact logic in App.tsx; if the production code
 * diverges these tests serve as the regression anchor.
 */

import { describe, expect, it } from 'vitest';
import {
  INITIAL_CAT_DATA,
  INITIAL_TAX_DATA,
  INITIAL_TREE,
  RETAIL_PLANS,
} from '../../src/data/mockData';
import type { Category, Metric, Plan, Taxonomy, TreeNode } from '../../src/types';

// ---------------------------------------------------------------------------
// Extracted pure-functions (mirroring App.tsx handlers)
// ---------------------------------------------------------------------------

// ---- Category handlers ----

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function addCat(categories: Category[], data: any): Category[] {
  const newCat: Category = {
    id: 'cat-' + data.key,
    _children: [],
    _locales: [],
    _props: [],
    ...data,
  };
  return [...categories, newCat];
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function editCat(categories: Category[], catId: string, data: any): Category[] {
  return categories.map(c => (c.id === catId ? { ...c, ...data } : c));
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function addEl(categories: Category[], catId: string, data: any): Category[] {
  return categories.map(c => {
    if (c.id === catId) {
      return { ...c, _children: [...c._children, { id: 'el-' + data.key, ...data }] };
    }
    return c;
  });
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function editEl(categories: Category[], catId: string, elId: string, data: any): Category[] {
  return categories.map(c => {
    if (c.id === catId) {
      return { ...c, _children: c._children.map(el => (el.id === elId ? { ...el, ...data } : el)) };
    }
    return c;
  });
}

function deleteCat(categories: Category[], catId: string): Category[] {
  return categories.filter(c => c.id !== catId);
}

function deleteEl(categories: Category[], catId: string, elId: string): Category[] {
  return categories.map(c => {
    if (c.id === catId) {
      return { ...c, _children: c._children.filter(el => el.id !== elId) };
    }
    return c;
  });
}

// ---- Taxonomy handlers ----

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function addTax(taxonomies: Taxonomy[], data: any): Taxonomy[] {
  return [...taxonomies, { id: 'tax-' + data.key, _props: [], _locales: [], dag: [], ...data }];
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function editTax(taxonomies: Taxonomy[], taxId: string, data: any): Taxonomy[] {
  return taxonomies.map(t => (t.id === taxId ? { ...t, ...data } : t));
}

function deleteTax(taxonomies: Taxonomy[], taxId: string): Taxonomy[] {
  return taxonomies.filter(t => t.id !== taxId);
}

// ---- Metric handlers ----

type PlanMetrics = Record<number, Metric[]>;

function addMetric(
  planMetrics: PlanMetrics,
  planId: number,
  currentMetrics: Metric[],
  metric: Metric
): PlanMetrics {
  return { ...planMetrics, [planId]: [...currentMetrics, metric] };
}

function editMetric(
  planMetrics: PlanMetrics,
  planId: number,
  currentMetrics: Metric[],
  metric: Metric
): PlanMetrics {
  return { ...planMetrics, [planId]: currentMetrics.map(m => (m.id === metric.id ? metric : m)) };
}

function deleteMetric(
  planMetrics: PlanMetrics,
  planId: number,
  currentMetrics: Metric[],
  id: number
): PlanMetrics {
  return { ...planMetrics, [planId]: currentMetrics.filter(m => m.id !== id) };
}

function reorderMetric(
  planMetrics: PlanMetrics,
  planId: number,
  currentMetrics: Metric[],
  id: number,
  direction: 'up' | 'down'
): PlanMetrics {
  const arr = [...currentMetrics];
  const i = arr.findIndex(m => m.id === id);
  if (direction === 'up' && i > 0) {
    [arr[i - 1], arr[i]] = [arr[i], arr[i - 1]];
  } else if (direction === 'down' && i < arr.length - 1) {
    [arr[i + 1], arr[i]] = [arr[i], arr[i + 1]];
  }
  return { ...planMetrics, [planId]: arr };
}

// ---- Plan handlers ----

function updatePlansInTree(
  nodes: TreeNode[],
  nodeId: number,
  updater: (plans: Plan[]) => Plan[]
): TreeNode[] {
  return nodes.map(node => {
    const newNode = { ...node };
    if (node.id === nodeId) {
      newNode.plans = updater(node.plans || []);
    }
    if (node.children) {
      newNode.children = updatePlansInTree(node.children, nodeId, updater);
    }
    return newNode;
  });
}

function addPlan(tree: TreeNode[], nodeId: number, data: Partial<Plan>): TreeNode[] {
  const newPlan: Plan = {
    id: 99999, // deterministic for tests
    key: data.key || 'NEW',
    label: data.label || 'New Plan',
    kind: data.kind || 'report',
    tooltip: data.tooltip || '',
    desc: data.desc || '',
  };
  return updatePlansInTree(tree, nodeId, plans => [...plans, newPlan]);
}

function editPlan(
  tree: TreeNode[],
  nodeId: number,
  planId: number,
  data: Partial<Plan>
): TreeNode[] {
  return updatePlansInTree(tree, nodeId, plans =>
    plans.map(p => (p.id === planId ? { ...p, ...data } : p))
  );
}

function deletePlan(tree: TreeNode[], nodeId: number, planId: number): TreeNode[] {
  return updatePlansInTree(tree, nodeId, plans => plans.filter(p => p.id !== planId));
}

// ---- Expand / collapse ----

function toggleExpand(ids: string[], id: string, expanded: boolean): string[] {
  if (expanded) return ids.includes(id) ? ids : [...ids, id];
  return ids.filter(x => x !== id);
}

// ---------------------------------------------------------------------------
// Helpers for tests
// ---------------------------------------------------------------------------

function findNodeById(nodes: TreeNode[], id: number): TreeNode | null {
  for (const n of nodes) {
    if (n.id === id) return n;
    if (n.children) {
      const f = findNodeById(n.children, id);
      if (f) return f;
    }
  }
  return null;
}

const BASE_CATS = () => JSON.parse(JSON.stringify(INITIAL_CAT_DATA)) as Category[];
const BASE_TAXES = () => JSON.parse(JSON.stringify(INITIAL_TAX_DATA)) as Taxonomy[];
const BASE_TREE = () => JSON.parse(JSON.stringify(INITIAL_TREE)) as TreeNode[];

// ---------------------------------------------------------------------------
// Category tests
// ---------------------------------------------------------------------------

describe('addCat', () => {
  it('appends a new category to the list', () => {
    const result = addCat(BASE_CATS(), { key: 'regions', name: 'Regions' });
    expect(result).toHaveLength(INITIAL_CAT_DATA.length + 1);
  });

  it('new category id uses "cat-" prefix + key', () => {
    const result = addCat(BASE_CATS(), { key: 'regions', name: 'Regions' });
    const added = result[result.length - 1];
    expect(added.id).toBe('cat-regions');
  });

  it('new category initialises with empty _children, _locales, _props', () => {
    const result = addCat(BASE_CATS(), { key: 'regions', name: 'Regions' });
    const added = result[result.length - 1];
    expect(added._children).toEqual([]);
    expect(added._locales).toEqual([]);
    expect(added._props).toEqual([]);
  });

  it('does not mutate the original array', () => {
    const original = BASE_CATS();
    addCat(original, { key: 'x', name: 'X' });
    expect(original).toHaveLength(INITIAL_CAT_DATA.length);
  });
});

describe('editCat', () => {
  it('updates only the matching category', () => {
    const result = editCat(BASE_CATS(), 'cat-channel', { name: 'Updated Channel' });
    const updated = result.find(c => c.id === 'cat-channel')!;
    expect(updated.name).toBe('Updated Channel');
  });

  it('does not modify other categories', () => {
    const original = BASE_CATS();
    const result = editCat(original, 'cat-channel', { name: 'Updated' });
    const unchanged = result.find(c => c.id === 'cat-prodclass')!;
    const originalUnchanged = original.find(c => c.id === 'cat-prodclass')!;
    expect(unchanged.name).toBe(originalUnchanged.name);
  });

  it('list length stays the same', () => {
    const result = editCat(BASE_CATS(), 'cat-channel', { name: 'X' });
    expect(result).toHaveLength(INITIAL_CAT_DATA.length);
  });
});

describe('addEl', () => {
  it('adds an element to the correct category', () => {
    const result = addEl(BASE_CATS(), 'cat-channel', { key: 'marketplace', name: 'Marketplace' });
    const cat = result.find(c => c.id === 'cat-channel')!;
    expect(cat._children).toHaveLength(
      INITIAL_CAT_DATA.find(c => c.id === 'cat-channel')!._children.length + 1
    );
  });

  it('new element id uses "el-" prefix + key', () => {
    const result = addEl(BASE_CATS(), 'cat-channel', { key: 'marketplace', name: 'Marketplace' });
    const cat = result.find(c => c.id === 'cat-channel')!;
    const added = cat._children[cat._children.length - 1];
    expect(added.id).toBe('el-marketplace');
  });

  it('does not affect other categories', () => {
    const result = addEl(BASE_CATS(), 'cat-channel', { key: 'x', name: 'X' });
    const other = result.find(c => c.id === 'cat-prodclass')!;
    expect(other._children).toHaveLength(
      INITIAL_CAT_DATA.find(c => c.id === 'cat-prodclass')!._children.length
    );
  });
});

describe('editEl', () => {
  it('updates the named element in the correct category', () => {
    const result = editEl(BASE_CATS(), 'cat-channel', 'el-online', { name: 'Web' });
    const cat = result.find(c => c.id === 'cat-channel')!;
    const el = cat._children.find(e => e.id === 'el-online')!;
    expect(el.name).toBe('Web');
  });

  it('does not modify other elements in the same category', () => {
    const original = BASE_CATS();
    const result = editEl(original, 'cat-channel', 'el-online', { name: 'Web' });
    const cat = result.find(c => c.id === 'cat-channel')!;
    const instoreName = cat._children.find(e => e.id === 'el-instore')!.name;
    const originalName = original
      .find(c => c.id === 'cat-channel')!
      ._children.find(e => e.id === 'el-instore')!.name;
    expect(instoreName).toBe(originalName);
  });
});

describe('deleteCat', () => {
  it('removes the specified category', () => {
    const result = deleteCat(BASE_CATS(), 'cat-channel');
    expect(result.find(c => c.id === 'cat-channel')).toBeUndefined();
  });

  it('reduces the list length by 1', () => {
    const result = deleteCat(BASE_CATS(), 'cat-channel');
    expect(result).toHaveLength(INITIAL_CAT_DATA.length - 1);
  });

  it('does not remove unrelated categories', () => {
    const result = deleteCat(BASE_CATS(), 'cat-channel');
    expect(result.find(c => c.id === 'cat-prodclass')).toBeDefined();
  });
});

describe('deleteEl', () => {
  it('removes the element from the correct category', () => {
    const result = deleteEl(BASE_CATS(), 'cat-channel', 'el-online');
    const cat = result.find(c => c.id === 'cat-channel')!;
    expect(cat._children.find(e => e.id === 'el-online')).toBeUndefined();
  });

  it('does not remove elements from other categories', () => {
    const result = deleteEl(BASE_CATS(), 'cat-channel', 'el-online');
    const pc = result.find(c => c.id === 'cat-prodclass')!;
    expect(pc._children).toHaveLength(
      INITIAL_CAT_DATA.find(c => c.id === 'cat-prodclass')!._children.length
    );
  });
});

// ---------------------------------------------------------------------------
// Taxonomy tests
// ---------------------------------------------------------------------------

describe('addTax', () => {
  it('appends a new taxonomy to the list', () => {
    const result = addTax(BASE_TAXES(), { key: 'promo', name: 'Promotions' });
    expect(result).toHaveLength(INITIAL_TAX_DATA.length + 1);
  });

  it('new taxonomy id uses "tax-" prefix + key', () => {
    const result = addTax(BASE_TAXES(), { key: 'promo', name: 'Promotions' });
    const added = result[result.length - 1];
    expect(added.id).toBe('tax-promo');
  });

  it('new taxonomy initialises with empty dag, _props, _locales', () => {
    const result = addTax(BASE_TAXES(), { key: 'promo', name: 'Promotions' });
    const added = result[result.length - 1];
    expect(added.dag).toEqual([]);
    expect(added._props).toEqual([]);
    expect(added._locales).toEqual([]);
  });
});

describe('editTax', () => {
  it('updates only the matching taxonomy', () => {
    const result = editTax(BASE_TAXES(), 'tax-web-nav', { name: 'Updated Nav' });
    const updated = result.find(t => t.id === 'tax-web-nav')!;
    expect(updated.name).toBe('Updated Nav');
  });

  it('does not modify other taxonomies', () => {
    const original = BASE_TAXES();
    const result = editTax(original, 'tax-web-nav', { name: 'X' });
    const other = result.find(t => t.id === 'tax-inv')!;
    expect(other.name).toBe(original.find(t => t.id === 'tax-inv')!.name);
  });
});

describe('deleteTax', () => {
  it('removes the specified taxonomy', () => {
    const result = deleteTax(BASE_TAXES(), 'tax-web-nav');
    expect(result.find(t => t.id === 'tax-web-nav')).toBeUndefined();
  });

  it('reduces the list length by 1', () => {
    const result = deleteTax(BASE_TAXES(), 'tax-web-nav');
    expect(result).toHaveLength(INITIAL_TAX_DATA.length - 1);
  });
});

// ---------------------------------------------------------------------------
// Metric tests
// ---------------------------------------------------------------------------

const SAMPLE_METRICS: Metric[] = [
  { id: 1, measure: 'inventory', dims: ['country'], grain: 'daily' },
  { id: 2, measure: 'demand', dims: ['region'], grain: 'weekly' },
  { id: 3, measure: 'revenue', dims: ['channel'], grain: 'monthly' },
];

const NEW_METRIC: Metric = { id: 4, measure: 'cost', dims: ['category'], grain: 'quarterly' };

describe('addMetric', () => {
  it('appends the metric to the plan', () => {
    const result = addMetric({}, 1001, SAMPLE_METRICS, NEW_METRIC);
    expect(result[1001]).toHaveLength(4);
    expect(result[1001][3]).toEqual(NEW_METRIC);
  });

  it('does not mutate other plans', () => {
    const existing: PlanMetrics = { 9999: [SAMPLE_METRICS[0]] };
    const result = addMetric(existing, 1001, SAMPLE_METRICS, NEW_METRIC);
    expect(result[9999]).toHaveLength(1);
  });
});

describe('editMetric', () => {
  it('replaces the metric with the matching id', () => {
    const updated = { ...SAMPLE_METRICS[1], grain: 'monthly' as const };
    const result = editMetric({}, 1001, SAMPLE_METRICS, updated);
    expect(result[1001][1].grain).toBe('monthly');
  });

  it('does not touch metrics with different ids', () => {
    const updated = { ...SAMPLE_METRICS[1], grain: 'monthly' as const };
    const result = editMetric({}, 1001, SAMPLE_METRICS, updated);
    expect(result[1001][0]).toEqual(SAMPLE_METRICS[0]);
    expect(result[1001][2]).toEqual(SAMPLE_METRICS[2]);
  });
});

describe('deleteMetric', () => {
  it('removes the metric with the given id', () => {
    const result = deleteMetric({}, 1001, SAMPLE_METRICS, 2);
    expect(result[1001]).toHaveLength(2);
    expect(result[1001].find(m => m.id === 2)).toBeUndefined();
  });
});

describe('reorderMetric', () => {
  it('moves a metric up by swapping with the previous', () => {
    const result = reorderMetric({}, 1001, SAMPLE_METRICS, 2, 'up');
    expect(result[1001][0].id).toBe(2);
    expect(result[1001][1].id).toBe(1);
    expect(result[1001][2].id).toBe(3);
  });

  it('moves a metric down by swapping with the next', () => {
    const result = reorderMetric({}, 1001, SAMPLE_METRICS, 2, 'down');
    expect(result[1001][0].id).toBe(1);
    expect(result[1001][1].id).toBe(3);
    expect(result[1001][2].id).toBe(2);
  });

  it('no-op when moving the first metric up', () => {
    const result = reorderMetric({}, 1001, SAMPLE_METRICS, 1, 'up');
    expect(result[1001].map(m => m.id)).toEqual([1, 2, 3]);
  });

  it('no-op when moving the last metric down', () => {
    const result = reorderMetric({}, 1001, SAMPLE_METRICS, 3, 'down');
    expect(result[1001].map(m => m.id)).toEqual([1, 2, 3]);
  });
});

// ---------------------------------------------------------------------------
// Plan tests (tree mutations)
// ---------------------------------------------------------------------------

// Node ID of Retail / Production / Definition bucket = 1020
const DEFINITION_NODE_ID = 1020;

describe('addPlan', () => {
  it('adds a plan to the specified tree node', () => {
    const tree = BASE_TREE();
    const result = addPlan(tree, DEFINITION_NODE_ID, {
      key: 'q1_fct',
      label: 'Q1 Forecast',
      kind: 'forecast',
    });
    const node = findNodeById(result, DEFINITION_NODE_ID)!;
    expect(node.plans!).toHaveLength(RETAIL_PLANS.length + 1);
  });

  it('new plan has the provided label and kind', () => {
    const tree = BASE_TREE();
    const result = addPlan(tree, DEFINITION_NODE_ID, {
      key: 'q1',
      label: 'Q1 Forecast',
      kind: 'forecast',
    });
    const node = findNodeById(result, DEFINITION_NODE_ID)!;
    const added = node.plans![node.plans!.length - 1];
    expect(added.label).toBe('Q1 Forecast');
    expect(added.kind).toBe('forecast');
  });

  it('does not modify sibling bucket nodes', () => {
    const tree = BASE_TREE();
    const result = addPlan(tree, DEFINITION_NODE_ID, { key: 'x', label: 'X' });
    const planningNode = findNodeById(result, 1010); // Planning bucket
    // Planning node has its own plans copy, unchanged
    expect(planningNode?.plans).toHaveLength(RETAIL_PLANS.length);
  });
});

describe('editPlan', () => {
  it('updates the matching plan in the node', () => {
    const tree = BASE_TREE();
    const result = editPlan(tree, DEFINITION_NODE_ID, RETAIL_PLANS[0].id, { label: 'New Label' });
    const node = findNodeById(result, DEFINITION_NODE_ID)!;
    const edited = node.plans!.find(p => p.id === RETAIL_PLANS[0].id)!;
    expect(edited.label).toBe('New Label');
  });

  it('does not change other plans', () => {
    const tree = BASE_TREE();
    const result = editPlan(tree, DEFINITION_NODE_ID, RETAIL_PLANS[0].id, { label: 'X' });
    const node = findNodeById(result, DEFINITION_NODE_ID)!;
    expect(node.plans!.find(p => p.id === RETAIL_PLANS[1].id)!.label).toBe(RETAIL_PLANS[1].label);
  });
});

describe('deletePlan', () => {
  it('removes the specified plan from the node', () => {
    const tree = BASE_TREE();
    const result = deletePlan(tree, DEFINITION_NODE_ID, RETAIL_PLANS[0].id);
    const node = findNodeById(result, DEFINITION_NODE_ID)!;
    expect(node.plans!.find(p => p.id === RETAIL_PLANS[0].id)).toBeUndefined();
    expect(node.plans!).toHaveLength(RETAIL_PLANS.length - 1);
  });
});

// ---------------------------------------------------------------------------
// Expand / collapse toggle
// ---------------------------------------------------------------------------

describe('toggleExpand', () => {
  it('adds an id to the list when expanding', () => {
    const result = toggleExpand([], 'cat-channel', true);
    expect(result).toContain('cat-channel');
  });

  it('does not add a duplicate id when expanding an already-expanded item', () => {
    const result = toggleExpand(['cat-channel'], 'cat-channel', true);
    expect(result.filter(x => x === 'cat-channel')).toHaveLength(1);
  });

  it('removes an id from the list when collapsing', () => {
    const result = toggleExpand(['cat-channel', 'cat-prodclass'], 'cat-channel', false);
    expect(result).not.toContain('cat-channel');
    expect(result).toContain('cat-prodclass');
  });

  it('collapsing a non-expanded id is a no-op', () => {
    const result = toggleExpand(['cat-prodclass'], 'cat-channel', false);
    expect(result).toEqual(['cat-prodclass']);
  });
});
