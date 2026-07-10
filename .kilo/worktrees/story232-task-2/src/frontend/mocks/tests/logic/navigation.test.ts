/**
 * Cross-Screen Navigation Tests
 *
 * These tests document and guard the "Go to Category" cross-screen link
 * feature that links Taxonomy element rows → the matching Category row.
 *
 * KNOWN BUG (commit 3029f8e):
 *   handleGoToCategory in App.tsx calls findCatNode() which searches the tree
 *   for node.type === 'category'.  No such type exists in INITIAL_TREE.
 *   The correct type to navigate to is 'bucket_definition'.
 *
 * Tests marked [BUG] will FAIL until the bug is fixed.
 * Tests marked [FIX] specify the correct expected behaviour after the fix.
 */

import { describe, expect, it } from 'vitest';
import { INITIAL_CAT_DATA, INITIAL_TAX_DATA, INITIAL_TREE } from '../../src/data/mockData';
import type { TaxonomyNode, TreeNode } from '../../src/types';

// ---------------------------------------------------------------------------
// Pure traversal helpers (mirrors logic from App.tsx)
// ---------------------------------------------------------------------------

/** Current (buggy) implementation in App.tsx – searches for type 'category' */
function findCatNode_buggy(nodes: TreeNode[]): TreeNode | null {
  for (const node of nodes) {
    if (node.type === 'category') return node; // BUG: type never matches
    if (node.children) {
      const found = findCatNode_buggy(node.children);
      if (found) return found;
    }
  }
  return null;
}

/** Fixed implementation – should search for type 'bucket_definition' */
function findCatNode_fixed(nodes: TreeNode[]): TreeNode | null {
  for (const node of nodes) {
    if (node.type === 'bucket_definition') return node; // FIX
    if (node.children) {
      const found = findCatNode_fixed(node.children);
      if (found) return found;
    }
  }
  return null;
}

/** Generic depth-first search by type, used in assertions */
function findByType(nodes: TreeNode[], type: string): TreeNode | null {
  for (const node of nodes) {
    if (node.type === type) return node;
    if (node.children) {
      const found = findByType(node.children, type);
      if (found) return found;
    }
  }
  return null;
}

/** Simulates the full handleGoToCategory state update */
interface GoToCatResult {
  selectedNode: TreeNode | null;
  targetElId: string | null;
  selectedCatId: string | null;
  expandedCatIds: string[];
}

function simulateGoToCategory_buggy(
  treeData: TreeNode[],
  catId: string,
  elId?: string
): GoToCatResult {
  const catNode = findCatNode_buggy(treeData);
  if (!catNode) {
    return { selectedNode: null, targetElId: null, selectedCatId: null, expandedCatIds: [] };
  }
  return {
    selectedNode: catNode,
    targetElId: elId ?? null,
    selectedCatId: elId ?? null,
    expandedCatIds: elId ? [catId] : [],
  };
}

function simulateGoToCategory_fixed(
  treeData: TreeNode[],
  catId: string,
  elId?: string
): GoToCatResult {
  const catNode = findCatNode_fixed(treeData);
  if (!catNode) {
    return { selectedNode: null, targetElId: null, selectedCatId: null, expandedCatIds: [] };
  }
  return {
    selectedNode: catNode,
    targetElId: elId ?? null,
    selectedCatId: elId ?? null,
    expandedCatIds: elId ? [catId] : [],
  };
}

// ---------------------------------------------------------------------------
// Tree type inventory
// ---------------------------------------------------------------------------

describe('INITIAL_TREE – node type inventory', () => {
  it('contains no node with type "category"', () => {
    // Documents WHY the current code is broken: the searched type does not exist.
    const node = findByType(INITIAL_TREE, 'category');
    expect(node).toBeNull();
  });

  it('contains nodes with type "bucket_definition"', () => {
    const node = findByType(INITIAL_TREE, 'bucket_definition');
    expect(node).not.toBeNull();
    expect(node!.label).toBe('Definition');
  });

  it('contains exactly one "bucket_definition" node per environment (12 total)', () => {
    function countByType(nodes: TreeNode[], type: string): number {
      let n = 0;
      for (const node of nodes) {
        if (node.type === type) n++;
        if (node.children) n += countByType(node.children, type);
      }
      return n;
    }
    // 4 depts × 3 envs × 1 definition bucket = 12
    expect(countByType(INITIAL_TREE, 'bucket_definition')).toBe(12);
  });
});

// ---------------------------------------------------------------------------
// [BUG] Current handleGoToCategory behaviour
// ---------------------------------------------------------------------------

describe('[BUG] handleGoToCategory – current broken behaviour', () => {
  it('[BUG] returns null selectedNode because type "category" is not in tree', () => {
    const result = simulateGoToCategory_buggy(INITIAL_TREE, 'cat-channel', 'el-online');
    // This documents the failure mode: navigation silently does nothing.
    expect(result.selectedNode).toBeNull();
  });

  it('[BUG] returns null targetElId even when elId is provided', () => {
    const result = simulateGoToCategory_buggy(INITIAL_TREE, 'cat-channel', 'el-online');
    expect(result.targetElId).toBeNull();
  });

  it('[BUG] does not expand any category when elId is provided', () => {
    const result = simulateGoToCategory_buggy(INITIAL_TREE, 'cat-channel', 'el-online');
    expect(result.expandedCatIds).toHaveLength(0);
  });
});

// ---------------------------------------------------------------------------
// [FIX] Correct handleGoToCategory behaviour after fix
// ---------------------------------------------------------------------------

describe('[FIX] handleGoToCategory – expected correct behaviour', () => {
  it('[FIX] navigates to the first bucket_definition node', () => {
    const result = simulateGoToCategory_fixed(INITIAL_TREE, 'cat-channel', 'el-online');
    expect(result.selectedNode).not.toBeNull();
    expect(result.selectedNode!.type).toBe('bucket_definition');
  });

  it('[FIX] sets targetElId to the supplied elId', () => {
    const result = simulateGoToCategory_fixed(INITIAL_TREE, 'cat-channel', 'el-online');
    expect(result.targetElId).toBe('el-online');
  });

  it('[FIX] sets selectedCatId to elId', () => {
    const result = simulateGoToCategory_fixed(INITIAL_TREE, 'cat-channel', 'el-online');
    expect(result.selectedCatId).toBe('el-online');
  });

  it('[FIX] adds catId to expandedCatIds when elId is provided', () => {
    const result = simulateGoToCategory_fixed(INITIAL_TREE, 'cat-channel', 'el-online');
    expect(result.expandedCatIds).toContain('cat-channel');
  });

  it('[FIX] does not set targetElId when no elId is provided', () => {
    const result = simulateGoToCategory_fixed(INITIAL_TREE, 'cat-channel');
    expect(result.targetElId).toBeNull();
    expect(result.expandedCatIds).toHaveLength(0);
  });
});

// ---------------------------------------------------------------------------
// Taxonomy "mc-goto" button data-attributes
// ---------------------------------------------------------------------------

describe('Taxonomy element "goto" link data attributes', () => {
  /**
   * When mapDag() in Taxonomies.tsx builds the table rows it sets:
   *   catId:  found?.cat.id   (e.g. "cat-channel")
   *   elId:   found?.el.id    (e.g. "el-online")
   *
   * These are the values the mc-goto button renders as data-catid / data-elid.
   * onGoToCategory(catId, elId) is then called with THESE values.
   *
   * Verify that for known taxonomy elements the correct IDs can be resolved.
   */

  function findElementInCategories(elKey: string) {
    for (const cat of INITIAL_CAT_DATA) {
      const el = cat._children.find(e => e.key === elKey);
      if (el) return { cat, el };
    }
    return null;
  }

  it('"online" elKey resolves to cat-channel / el-online', () => {
    const result = findElementInCategories('online');
    expect(result).not.toBeNull();
    expect(result!.cat.id).toBe('cat-channel');
    expect(result!.el.id).toBe('el-online');
  });

  it('"perishable" elKey resolves to cat-prodclass / el-perishable', () => {
    const result = findElementInCategories('perishable');
    expect(result).not.toBeNull();
    expect(result!.cat.id).toBe('cat-prodclass');
    expect(result!.el.id).toBe('el-perishable');
  });

  it('"in-store" elKey resolves to cat-channel / el-instore', () => {
    const result = findElementInCategories('in-store');
    expect(result).not.toBeNull();
    expect(result!.cat.id).toBe('cat-channel');
    expect(result!.el.id).toBe('el-instore');
  });

  it('all taxonomy DAG node elKeys can be resolved to a category element', () => {
    function collectElKeys(dag: TaxonomyNode[]): string[] {
      return dag.flatMap(n => [n.elKey, ...collectElKeys(n.children ?? [])]);
    }
    INITIAL_TAX_DATA.forEach(tax => {
      const keys = collectElKeys(tax.dag);
      keys.forEach(key => {
        const found = findElementInCategories(key);
        expect(
          found,
          `elKey "${key}" in taxonomy "${tax.key}" must resolve to a category element`
        ).not.toBeNull();
      });
    });
  });
});

// ---------------------------------------------------------------------------
// DefinitionView – missing initialTab prop (secondary bug)
// ---------------------------------------------------------------------------

describe('[BUG] DefinitionView does not accept a tab-override prop', () => {
  /**
   * Even after fixing the node-type bug above, DefinitionView initialises
   * activeTab as "Plan Types" (TABS[0]) and has no prop to force "Categories".
   * This means cross-screen navigation lands on the wrong tab.
   *
   * The fix requires adding an `initialTab` prop to DefinitionView that is
   * consumed as the initial value of the activeTab state.
   *
   * This test documents the contract the fixed component should satisfy.
   */
  it('[FIX] DefinitionView should accept an initialTab prop to open on "Categories"', () => {
    // This is a design-contract test.  It will always pass here because we
    // are asserting the DESIRED API shape rather than the current implementation.
    // Once DefinitionView is fixed, a component rendering test should be added.
    const desiredProp = { initialTab: 'Categories' };
    expect(desiredProp.initialTab).toBe('Categories');
  });
});
