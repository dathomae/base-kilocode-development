import { describe, expect, it } from 'vitest';
import {
  INITIAL_CAT_DATA,
  INITIAL_CONVERTER_DATA,
  INITIAL_MEASURE_DATA,
  INITIAL_TAX_DATA,
  INITIAL_TREE,
  RETAIL_PLANS,
  makeElementShape,
  mkLoc,
  mkProp,
} from '../../src/data/mockData';
import type { Converter, Measure, TreeNode } from '../../src/types';

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function allNodes(nodes: TreeNode[]): TreeNode[] {
  return nodes.flatMap(n => [n, ...(n.children ? allNodes(n.children) : [])]);
}

function findByType(nodes: TreeNode[], type: string): TreeNode[] {
  return allNodes(nodes).filter(n => n.type === type);
}

// ---------------------------------------------------------------------------
// Tree structure
// ---------------------------------------------------------------------------

describe('INITIAL_TREE – structure', () => {
  it('has exactly one root node', () => {
    expect(INITIAL_TREE).toHaveLength(1);
    expect(INITIAL_TREE[0].type).toBe('root');
  });

  it('root has exactly 4 department children', () => {
    const depts = INITIAL_TREE[0].children ?? [];
    expect(depts).toHaveLength(4);
    depts.forEach(d => expect(d.type).toBe('dept'));
  });

  it('every department has exactly 3 environment children', () => {
    const depts = INITIAL_TREE[0].children ?? [];
    depts.forEach(dept => {
      const envs = dept.children ?? [];
      expect(envs).toHaveLength(3);
      envs.forEach(e => expect(e.type).toBe('env'));
    });
  });

  it('every environment has exactly 4 bucket children', () => {
    const envs = findByType(INITIAL_TREE, 'env');
    envs.forEach(env => {
      const buckets = env.children ?? [];
      expect(buckets).toHaveLength(4);
    });
  });

  it('bucket types are planning, definition, calculation, integration', () => {
    const envs = findByType(INITIAL_TREE, 'env');
    envs.forEach(env => {
      const types = (env.children ?? []).map(b => b.type);
      expect(types).toContain('bucket_planning');
      expect(types).toContain('bucket_definition');
      expect(types).toContain('bucket_calculation');
      expect(types).toContain('bucket_integration');
    });
  });

  it('all node IDs within the tree are unique', () => {
    const ids = allNodes(INITIAL_TREE).map(n => n.id);
    expect(new Set(ids).size).toBe(ids.length);
  });
});

// ---------------------------------------------------------------------------
// Plan data on Retail / Production buckets
// ---------------------------------------------------------------------------

describe('INITIAL_TREE – Retail Production plans', () => {
  function retailProduction(): TreeNode {
    const retail = (INITIAL_TREE[0].children ?? []).find(d => d.label === 'Retail')!;
    return (retail.children ?? []).find(e => e.label === 'Production')!;
  }

  it('Retail Production Planning bucket carries RETAIL_PLANS', () => {
    const planning = (retailProduction().children ?? []).find(b => b.type === 'bucket_planning')!;
    expect(planning.plans).toEqual(RETAIL_PLANS);
  });

  it('Retail Production Definition bucket carries RETAIL_PLANS', () => {
    const def = (retailProduction().children ?? []).find(b => b.type === 'bucket_definition')!;
    expect(def.plans).toEqual(RETAIL_PLANS);
  });

  it('Staging and Development environments have no plans by default', () => {
    const retail = (INITIAL_TREE[0].children ?? []).find(d => d.label === 'Retail')!;
    const staging = (retail.children ?? []).find(e => e.label === 'Staging')!;
    const dev = (retail.children ?? []).find(e => e.label === 'Development')!;
    [staging, dev].forEach(env => {
      (env.children ?? []).forEach(b => {
        expect(b.plans ?? []).toHaveLength(0);
      });
    });
  });
});

// ---------------------------------------------------------------------------
// RETAIL_PLANS
// ---------------------------------------------------------------------------

describe('RETAIL_PLANS', () => {
  it('has exactly 7 plan entries', () => {
    expect(RETAIL_PLANS).toHaveLength(7);
  });

  it('has 4 reports and 3 forecasts', () => {
    const reports = RETAIL_PLANS.filter(p => p.kind === 'report');
    const forecasts = RETAIL_PLANS.filter(p => p.kind === 'forecast');
    expect(reports).toHaveLength(4);
    expect(forecasts).toHaveLength(3);
  });

  it('all plan IDs are unique', () => {
    const ids = RETAIL_PLANS.map(p => p.id);
    expect(new Set(ids).size).toBe(ids.length);
  });

  it('all plan keys are unique', () => {
    const keys = RETAIL_PLANS.map(p => p.key);
    expect(new Set(keys).size).toBe(keys.length);
  });
});

// ---------------------------------------------------------------------------
// INITIAL_CAT_DATA
// ---------------------------------------------------------------------------

describe('INITIAL_CAT_DATA', () => {
  it('has exactly 10 categories', () => {
    expect(INITIAL_CAT_DATA).toHaveLength(10);
  });

  it('all category IDs are unique', () => {
    const ids = INITIAL_CAT_DATA.map(c => c.id);
    expect(new Set(ids).size).toBe(ids.length);
  });

  it('all category element IDs are globally unique', () => {
    const elIds = INITIAL_CAT_DATA.flatMap(c => c._children.map(e => e.id));
    expect(new Set(elIds).size).toBe(elIds.length);
  });

  it('"channel" category has exactly 3 elements', () => {
    const ch = INITIAL_CAT_DATA.find(c => c.key === 'channel')!;
    expect(ch).toBeDefined();
    expect(ch._children).toHaveLength(3);
  });

  it('"product_class" category has exactly 4 elements', () => {
    const pc = INITIAL_CAT_DATA.find(c => c.key === 'product_class')!;
    expect(pc).toBeDefined();
    expect(pc._children).toHaveLength(4);
  });

  it('"data-types" category has exactly 4 elements', () => {
    const dt = INITIAL_CAT_DATA.find(c => c.key === 'data-types')!;
    expect(dt).toBeDefined();
    expect(dt._children).toHaveLength(4);
  });

  it('each category object has required fields', () => {
    INITIAL_CAT_DATA.forEach(c => {
      expect(c).toHaveProperty('id');
      expect(c).toHaveProperty('key');
      expect(c).toHaveProperty('name');
      expect(c).toHaveProperty('_children');
      expect(c).toHaveProperty('_locales');
      expect(c).toHaveProperty('_props');
    });
  });
});

// ---------------------------------------------------------------------------
// INITIAL_TAX_DATA
// ---------------------------------------------------------------------------

describe('INITIAL_TAX_DATA', () => {
  it('has exactly 2 taxonomies', () => {
    expect(INITIAL_TAX_DATA).toHaveLength(2);
  });

  it('"inventory-locations" taxonomy has isMeasuresDimension = true', () => {
    const inv = INITIAL_TAX_DATA.find(t => t.key === 'inventory-locations')!;
    expect(inv).toBeDefined();
    expect(inv.isMeasuresDimension).toBe(true);
  });

  it('"web-nav" taxonomy does NOT have isMeasuresDimension', () => {
    const wn = INITIAL_TAX_DATA.find(t => t.key === 'web-nav')!;
    expect(wn).toBeDefined();
    expect(wn.isMeasuresDimension).toBeFalsy();
  });

  it('"web-nav" allowedCats includes channel and product_class', () => {
    const wn = INITIAL_TAX_DATA.find(t => t.key === 'web-nav')!;
    expect(wn.allowedCats).toContain('channel');
    expect(wn.allowedCats).toContain('product_class');
  });

  it('each taxonomy has a non-empty DAG', () => {
    INITIAL_TAX_DATA.forEach(t => {
      expect(t.dag.length).toBeGreaterThan(0);
    });
  });
});

// ---------------------------------------------------------------------------
// INITIAL_MEASURE_DATA
// ---------------------------------------------------------------------------

describe('INITIAL_MEASURE_DATA', () => {
  it('has exactly 3 measures', () => {
    expect(INITIAL_MEASURE_DATA).toHaveLength(3);
  });

  it('all measure IDs are unique', () => {
    const ids = INITIAL_MEASURE_DATA.map((m: Measure) => m.id);

    expect(new Set(ids).size).toBe(ids.length);
  });

  it('each measure has id, key, name, dataType, units', () => {
    INITIAL_MEASURE_DATA.forEach((m: Measure) => {
      expect(m).toHaveProperty('id');
      expect(m).toHaveProperty('key');
      expect(m).toHaveProperty('name');
      expect(m).toHaveProperty('dataType');
      expect(m).toHaveProperty('units');
    });
  });
});

// ---------------------------------------------------------------------------
// INITIAL_CONVERTER_DATA
// ---------------------------------------------------------------------------

describe('INITIAL_CONVERTER_DATA', () => {
  it('has exactly 2 converters', () => {
    expect(INITIAL_CONVERTER_DATA).toHaveLength(2);
  });

  it('converter types are Measure or Unit', () => {
    INITIAL_CONVERTER_DATA.forEach((c: Converter) => {
      expect(['Measure', 'Unit']).toContain(c.type);
    });
  });

  it('each converter has assignment, input, and calculation', () => {
    INITIAL_CONVERTER_DATA.forEach((c: Converter) => {
      expect(c).toHaveProperty('assignment');
      expect(c).toHaveProperty('input');
      expect(c).toHaveProperty('calculation');
    });
  });
});

// ---------------------------------------------------------------------------
// Factory helpers
// ---------------------------------------------------------------------------

describe('mkLoc', () => {
  it('creates a Locale with all required fields', () => {
    const loc = mkLoc('en', 'US', 'Name', 'Tooltip', 'Desc');
    expect(loc).toEqual({
      lang: 'en',
      region: 'US',
      name: 'Name',
      tooltip: 'Tooltip',
      desc: 'Desc',
    });
  });
});

describe('mkProp', () => {
  it('creates a Property with correct shape', () => {
    const prop = mkProp('my_key', 'string', 'hello');
    expect(prop).toMatchObject({ key: 'my_key', type: 'string', value: 'hello' });
    expect(prop.id).toBeTruthy();
  });
});

describe('makeElementShape', () => {
  it('creates an object-type property keyed "element-shape"', () => {
    const p = makeElementShape();
    expect(p.key).toBe('element-shape');
    expect(p.type).toBe('object');
  });

  it('value is valid JSON containing value-type and units', () => {
    const p = makeElementShape();
    expect(() => JSON.parse(p.value)).not.toThrow();
    const parsed = JSON.parse(p.value);
    expect(parsed).toHaveProperty('value-type');
    expect(parsed).toHaveProperty('units');
  });
});
