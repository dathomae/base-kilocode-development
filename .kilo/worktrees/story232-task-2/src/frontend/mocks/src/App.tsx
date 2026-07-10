import React, { useCallback, useState } from 'react';
import { Footer } from './components/Layout/Footer';
import { Header } from './components/Layout/Header';
import { Sidebar } from './components/Layout/Sidebar';
import { CategoryModal } from './components/Modals/CategoryModal';
import { MetricModal } from './components/Modals/MetricModal';
import { PlanModal } from './components/Modals/PlanModal';
import { TaxonomyElementModal } from './components/Modals/TaxonomyElementModal';
import { TaxonomyModal } from './components/Modals/TaxonomyModal';
import { CalculationView } from './components/Views/Calculation/CalculationView';
import { DefinitionView } from './components/Views/Definition/DefinitionView';
import { IntegrationView } from './components/Views/Integration/IntegrationView';
import { PlanningView } from './components/Views/Planning/PlanningView';
import {
  INITIAL_CAT_DATA,
  INITIAL_CONVERTER_DATA,
  INITIAL_MEASURE_DATA,
  INITIAL_TAX_DATA,
  INITIAL_TREE,
} from './data/mockData';
import {
  Category,
  Converter,
  Measure,
  Metric,
  Plan,
  Taxonomy,
  TaxonomyNode,
  TreeNode,
} from './types';

const HELP_MESSAGES: Record<string, string> = {
  root: 'Browsing the top-level department hierarchy.',
  dept: 'Select an environment — Production, Staging, or Development.',
  env: 'Select a governance category.',
  bucket_planning: 'Strategic planning and execution tracking.',
  bucket_definition: 'Core structural definitions and governance.',
  bucket_calculation: 'Quantitative measures and calculation logic.',
  bucket_integration: 'Data connectivity and mapping configurations.',
};

// Removed BUCKET_TABS as they are now managed within each bucket view

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

export default function App() {
  const [treeData, setTreeData] = useState<TreeNode[]>(INITIAL_TREE);
  const [selectedNode, setSelectedNode] = useState<TreeNode | null>(null);
  const [isSidebarCollapsed, setIsSidebarCollapsed] = useState(false);
  const [isHeaderMinimized, setIsHeaderMinimized] = useState(false);
  const [activePlanId, setActivePlanId] = useState<number | null>(null);
  // Removed activeTab state as it was not used or handled within subcomponents

  const [effectiveDate, setEffectiveDate] = useState('Now');
  const [planMetrics, setPlanMetrics] = useState<Record<number, Metric[]>>({});
  const [categories, setCategories] = useState<Category[]>(INITIAL_CAT_DATA);
  const [taxonomies, setTaxonomies] = useState<Taxonomy[]>(INITIAL_TAX_DATA);
  const [measures, setMeasures] = useState<Measure[]>(INITIAL_MEASURE_DATA);
  const [converters, setConverters] = useState<Converter[]>(INITIAL_CONVERTER_DATA);

  // Tree state
  const [expandedCatIds, setExpandedCatIds] = useState<string[]>([]);
  const [expandedTaxIds, setExpandedTaxIds] = useState<string[]>([]);
  const [selectedCatId, setSelectedCatId] = useState<string | null>(null);
  const [selectedTaxId, setSelectedTaxId] = useState<string | null>(null);
  const [selectedMetricId, setSelectedMetricId] = useState<string | null>(null);
  const [targetElId, setTargetElId] = useState<string | null>(null);

  // Modals state
  const [isMetricModalOpen, setIsMetricModalOpen] = useState(false);
  const [editingMetric, setEditingMetric] = useState<Metric | null>(null);

  const [isCatModalOpen, setIsCatModalOpen] = useState(false);
  const [catModalMode, setCatModalMode] = useState<'addCat' | 'editCat' | 'addEl' | 'editEl'>(
    'addCat'
  );
  const [editingCatId, setEditingCatId] = useState<string | null>(null);
  const [editingElId, setEditingElId] = useState<string | null>(null);
  const [initialCatTab, setInitialCatTab] = useState<'general' | 'properties' | 'locales'>(
    'general'
  );

  const [isTaxModalOpen, setIsTaxModalOpen] = useState(false);
  const [editingTaxonomy, setEditingTaxonomy] = useState<Taxonomy | null>(null);
  const [initialTaxTab, setInitialTaxTab] = useState<'general' | 'properties' | 'locales' | 'dag'>(
    'general'
  );

  const [isPlanModalOpen, setIsPlanModalOpen] = useState(false);
  const [editingPlan, setEditingPlan] = useState<Plan | null>(null);

  const handleNodeSelect = (node: TreeNode) => {
    setSelectedNode(node);
    setActivePlanId(null);
  };

  const handleDateChange = (val: number) => {
    if (val === 100) {
      setEffectiveDate('Now');
    } else {
      const T0 = new Date('2024-01-01').getTime();
      const T1 = Date.now();
      const date = new Date(T0 + (T1 - T0) * (val / 100));
      setEffectiveDate(
        date.toLocaleDateString('en-US', { year: '2-digit', month: 'short', day: 'numeric' })
      );
    }
  };

  const getMetricsForPlan = (planId: number, planLabel: string) => {
    return planMetrics[planId] || DEFAULT_METRICS(planId, planLabel);
  };

  const handleMetricConfirm = (metric: Metric) => {
    if (!activePlanId || !selectedNode) return;
    const plan = selectedNode.plans?.find(p => p.id === activePlanId);
    if (!plan) return;
    const current = getMetricsForPlan(activePlanId, plan.label);
    if (editingMetric) {
      setPlanMetrics(prev => ({
        ...prev,
        [activePlanId]: current.map(m => (m.id === metric.id ? metric : m)),
      }));
    } else {
      setPlanMetrics(prev => ({
        ...prev,
        [activePlanId]: [...current, metric],
      }));
    }
    setIsMetricModalOpen(false);
  };

  const handleMetricDelete = (id: number) => {
    if (!activePlanId || !selectedNode) return;
    const plan = selectedNode.plans?.find(p => p.id === activePlanId);
    if (!plan) return;
    const current = getMetricsForPlan(activePlanId, plan.label);
    setPlanMetrics(prev => ({
      ...prev,
      [activePlanId]: current.filter(m => m.id !== id),
    }));
    setIsMetricModalOpen(false);
  };

  const handleMetricReorder = (id: number, direction: 'up' | 'down') => {
    if (!activePlanId || !selectedNode) return;
    const plan = selectedNode.plans?.find(p => p.id === activePlanId);
    if (!plan) return;
    const arr = [...getMetricsForPlan(activePlanId, plan.label)];
    const i = arr.findIndex(m => m.id === id);
    if (direction === 'up' && i > 0) {
      [arr[i - 1], arr[i]] = [arr[i], arr[i - 1]];
    } else if (direction === 'down' && i < arr.length - 1) {
      [arr[i + 1], arr[i]] = [arr[i], arr[i + 1]];
    }
    setPlanMetrics(prev => ({ ...prev, [activePlanId]: arr }));
  };

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const handleCatConfirm = (data: any) => {
    if (catModalMode === 'addCat') {
      const newCat: Category = {
        id: 'cat-' + data.key,
        _children: [],
        _locales: [],
        _props: [],
        ...data,
      };
      setCategories([...categories, newCat]);
    } else if (catModalMode === 'editCat') {
      setCategories(categories.map(c => (c.id === editingCatId ? { ...c, ...data } : c)));
    } else if (catModalMode === 'addEl') {
      setCategories(
        categories.map(c => {
          if (c.id === editingCatId) {
            return { ...c, _children: [...c._children, { id: 'el-' + data.key, ...data }] };
          }
          return c;
        })
      );
    } else if (catModalMode === 'editEl') {
      setCategories(
        categories.map(c => {
          if (c.id === editingCatId) {
            const elId = editingElId || 'el-' + data.key;
            return {
              ...c,
              _children: c._children.map(el => (el.id === elId ? { ...el, ...data } : el)),
            };
          }
          return c;
        })
      );
    }
    setIsCatModalOpen(false);
  };

  const handleCatDelete = () => {
    if (catModalMode === 'editCat') {
      setCategories(categories.filter(c => c.id !== editingCatId));
    } else if (catModalMode === 'editEl') {
      setCategories(
        categories.map(c => {
          if (c.id === editingCatId) {
            return { ...c, _children: c._children.filter(el => el.id !== editingElId) };
          }
          return c;
        })
      );
    }
    setIsCatModalOpen(false);
  };

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const handleTaxConfirm = (data: any) => {
    if (editingTaxonomy) {
      setTaxonomies(taxonomies.map(t => (t.id === editingTaxonomy.id ? { ...t, ...data } : t)));
    } else {
      setTaxonomies([
        ...taxonomies,
        { id: 'tax-' + data.key, _props: [], _locales: [], dag: [], ...data },
      ]);
    }
    setIsTaxModalOpen(false);
  };

  const handleTaxDelete = useCallback(() => {
    if (editingTaxonomy) {
      setTaxonomies(taxonomies.filter(t => t.id !== editingTaxonomy.id));
    }
    setIsTaxModalOpen(false);
  }, [editingTaxonomy, taxonomies]);

  const handlePlanConfirm = (data: Partial<Plan>) => {
    if (!selectedNode) return;

    const updatePlansInTree = (nodes: TreeNode[]): TreeNode[] => {
      return nodes.map(node => {
        const newNode = { ...node };
        if (node.id === selectedNode.id) {
          const currentPlans = node.plans || [];
          if (editingPlan) {
            newNode.plans = currentPlans.map(p =>
              p.id === editingPlan.id ? { ...p, ...data } : p
            );
          } else {
            const newPlan: Plan = {
              id: Date.now(),
              key: data.key || 'NEW',
              label: data.label || 'New Plan',
              kind: data.kind || 'report',
              tooltip: data.tooltip || '',
              desc: data.desc || '',
            };
            newNode.plans = [...currentPlans, newPlan];
          }
        }
        if (node.children) {
          newNode.children = updatePlansInTree(node.children);
        }
        return newNode;
      });
    };

    const newTree = updatePlansInTree(treeData);
    setTreeData(newTree);

    // Also update selectedNode to reflect changes immediately
    const findNode = (nodes: TreeNode[]): TreeNode | null => {
      for (const n of nodes) {
        if (n.id === selectedNode.id) return n;
        if (n.children) {
          const found = findNode(n.children);
          if (found) return found;
        }
      }
      return null;
    };
    const updated = findNode(newTree);
    if (updated) {
      setSelectedNode(updated);
      if (!activePlanId && updated.plans && updated.plans.length > 0) {
        setActivePlanId(updated.plans[0].id);
      }
    }

    setIsPlanModalOpen(false);
  };

  const handlePlanDelete = (id: number) => {
    if (!selectedNode) return;

    const updatePlansInTree = (nodes: TreeNode[]): TreeNode[] => {
      return nodes.map(node => {
        const newNode = { ...node };
        if (node.id === selectedNode.id) {
          newNode.plans = (node.plans || []).filter(p => p.id !== id);
        }
        if (node.children) {
          newNode.children = updatePlansInTree(node.children);
        }
        return newNode;
      });
    };

    const newTree = updatePlansInTree(treeData);
    setTreeData(newTree);

    const findNode = (nodes: TreeNode[]): TreeNode | null => {
      for (const n of nodes) {
        if (n.id === selectedNode.id) return n;
        if (n.children) {
          const found = findNode(n.children);
          if (found) return found;
        }
      }
      return null;
    };
    const updated = findNode(newTree);
    if (updated) {
      setSelectedNode(updated);
      if (activePlanId === id) {
        setActivePlanId(updated.plans && updated.plans.length > 0 ? updated.plans[0].id : null);
      }
    }
  };

  const handleGoToCategory = useCallback(
    (catId: string, elId?: string) => {
      // Find the Categories node in the tree
      const findCatNode = (nodes: TreeNode[]): TreeNode | null => {
        for (const node of nodes) {
          if (node.type === 'category') return node;
          if (node.children) {
            const found = findCatNode(node.children);
            if (found) return found;
          }
        }
        return null;
      };

      const catNode = findCatNode(treeData);
      if (catNode) {
        setSelectedNode(catNode);
        if (elId) {
          setTargetElId(elId);
          setSelectedCatId(elId);
          // Ensure the category itself is expanded in the state
          setExpandedCatIds(prev => (prev.includes(catId) ? prev : [...prev, catId]));
        }
      }
    },
    [treeData]
  );

  const [isTaxElModalOpen, setIsTaxElModalOpen] = useState(false);
  const [activeTaxId, setActiveTaxId] = useState<string | null>(null);
  const [activeParentNodeId, setActiveParentNodeId] = useState<string | null>(null);

  const handleAddTaxElement = useCallback((taxId: string, parentNodeId?: string) => {
    setActiveTaxId(taxId);
    setActiveParentNodeId(parentNodeId || null);
    setIsTaxElModalOpen(true);
  }, []);

  const onTaxElConfirm = useCallback(
    (elKey: string) => {
      if (!activeTaxId) return;
      const newNode: TaxonomyNode = { id: 'tn-' + Date.now(), elKey, children: [] };

      setTaxonomies(prev =>
        prev.map(t => {
          if (t.id === activeTaxId) {
            if (!activeParentNodeId) {
              return { ...t, dag: [...t.dag, newNode] };
            } else {
              const updateNodes = (nodes: TaxonomyNode[]): TaxonomyNode[] => {
                return nodes.map(n => {
                  if (n.id === activeParentNodeId) {
                    return { ...n, children: [...n.children, newNode] };
                  }
                  return { ...n, children: updateNodes(n.children) };
                });
              };
              return { ...t, dag: updateNodes(t.dag) };
            }
          }
          return t;
        })
      );
      setIsTaxElModalOpen(false);
    },
    [activeTaxId, activeParentNodeId]
  );

  const handleToggleCatExpand = useCallback((id: string, expanded: boolean) => {
    setExpandedCatIds(prev =>
      expanded ? (prev.includes(id) ? prev : [...prev, id]) : prev.filter(x => x !== id)
    );
  }, []);

  const handleToggleTaxExpand = useCallback((id: string, expanded: boolean) => {
    setExpandedTaxIds(prev =>
      expanded ? (prev.includes(id) ? prev : [...prev, id]) : prev.filter(x => x !== id)
    );
  }, []);

  const handleAddElementToCat = useCallback((catId: string) => {
    setCatModalMode('addEl');
    setEditingCatId(catId);
    setEditingElId(null);
    setIsCatModalOpen(true);
  }, []);

  const handleEditCatOrEl = useCallback(
    (id: string, isCat: boolean, tab: 'general' | 'properties' | 'locales' = 'general') => {
      if (isCat) {
        setCatModalMode('editCat');
        setEditingCatId(id);
        setEditingElId(null);
      } else {
        setCatModalMode('editEl');
        // We need to find the category ID for this element
        // This is a bit expensive but fine for mock
        setCategories(prev => {
          const cat = prev.find(c => c._children.some(el => el.id === id));
          setEditingCatId(cat?.id || null);
          return prev;
        });
        setEditingElId(id);
      }
      setInitialCatTab(tab);
      setIsCatModalOpen(true);
    },
    []
  );

  const handleEditTax = useCallback(
    (id: string, tab: 'general' | 'properties' | 'locales' | 'dag' = 'general') => {
      setTaxonomies(prev => {
        setEditingTaxonomy(prev.find(t => t.id === id) || null);
        return prev;
      });
      setInitialTaxTab(tab);
      setIsTaxModalOpen(true);
    },
    []
  );

  const renderDetailContent = () => {
    if (!selectedNode) {
      return (
        <div className="placeholder text-center text-[var(--muted)] text-[15px] mt-20">
          <div className="icon text-5xl mb-3">🌲</div>
          <div>Select an item from the tree to view details</div>
        </div>
      );
    }

    const findParent = (
      nodes: TreeNode[],
      targetId: number,
      parent: TreeNode | null = null
    ): TreeNode | null => {
      for (const node of nodes) {
        if (node.id === targetId) return parent;
        if (node.children) {
          const found = findParent(node.children, targetId, node);
          if (found) return found;
        }
      }
      return null;
    };

    const getBreadcrumbs = (node: TreeNode): TreeNode[] => {
      const path: TreeNode[] = [node];
      let current = node;
      while (true) {
        const parent = findParent(treeData, current.id);
        if (!parent) break;
        path.unshift(parent);
        current = parent;
      }
      return path;
    };

    const breadcrumbs = getBreadcrumbs(selectedNode);
    const parentLabel = breadcrumbs.length > 1 ? breadcrumbs[breadcrumbs.length - 2].label : '';
    const bc = breadcrumbs.map((node, i) => (
      <React.Fragment key={node.id}>
        {i > 0 && <span className="mx-1.5 opacity-60">&gt;</span>}
        {i < breadcrumbs.length - 1 ? (
          <button
            onClick={() => handleNodeSelect(node)}
            className="hover:text-[var(--text)] transition-colors cursor-pointer border-none bg-none p-0 uppercase tracking-wider font-inherit text-[11px] text-[var(--muted)]"
          >
            {node.label}
          </button>
        ) : (
          <span>{node.label}</span>
        )}
      </React.Fragment>
    ));

    const title = selectedNode.label;

    const renderBucketView = (node: TreeNode) => {
      const helpMsg =
        HELP_MESSAGES[node.type as keyof typeof HELP_MESSAGES] ||
        'Manage your environment settings and configurations.';

      if (node.type === 'bucket_planning') {
        return (
          <PlanningView
            node={node}
            breadcrumbs={bc}
            helpMsg={helpMsg}
            parentLabel={parentLabel}
            isHeaderMinimized={isHeaderMinimized}
            onToggleHeader={() => setIsHeaderMinimized(!isHeaderMinimized)}
          />
        );
      }

      if (node.type === 'bucket_definition') {
        return (
          <DefinitionView
            node={node}
            breadcrumbs={bc}
            helpMsg={helpMsg}
            parentLabel={parentLabel}
            isHeaderMinimized={isHeaderMinimized}
            onToggleHeader={() => setIsHeaderMinimized(!isHeaderMinimized)}
            plans={node.plans || []}
            onAddPlan={() => {
              setEditingPlan(null);
              setIsPlanModalOpen(true);
            }}
            onEditPlan={p => {
              setEditingPlan(p);
              setIsPlanModalOpen(true);
            }}
            onDeletePlan={handlePlanDelete}
            getMetricsForPlan={getMetricsForPlan}
            onAddMetric={() => {
              setEditingMetric(null);
              setIsMetricModalOpen(true);
            }}
            onEditMetric={m => {
              setEditingMetric(m);
              setIsMetricModalOpen(true);
            }}
            onReorderMetric={handleMetricReorder}
            selectedMetricId={selectedMetricId}
            onSelectMetric={setSelectedMetricId}
            activePlanId={activePlanId}
            setActivePlanId={setActivePlanId}
            categories={categories}
            expandedCatIds={expandedCatIds}
            onToggleCatExpand={handleToggleCatExpand}
            selectedCatId={selectedCatId}
            onSelectCat={setSelectedCatId}
            targetElId={targetElId}
            onClearTarget={() => setTargetElId(null)}
            onAddCategory={() => {
              setCatModalMode('addCat');
              setEditingCatId(null);
              setIsCatModalOpen(true);
            }}
            onAddElementToCat={handleAddElementToCat}
            onEditCatOrEl={handleEditCatOrEl}
            taxonomies={taxonomies}
            expandedTaxIds={expandedTaxIds}
            onToggleTaxExpand={handleToggleTaxExpand}
            selectedTaxId={selectedTaxId}
            onSelectTax={setSelectedTaxId}
            onAddTaxonomy={() => {
              setEditingTaxonomy(null);
              setIsTaxModalOpen(true);
            }}
            onEditTaxonomy={handleEditTax}
            onAddTaxElement={handleAddTaxElement}
            onGoToCategory={handleGoToCategory}
          />
        );
      }

      if (node.type === 'bucket_calculation') {
        return (
          <CalculationView
            node={node}
            breadcrumbs={bc}
            helpMsg={helpMsg}
            parentLabel={parentLabel}
            isHeaderMinimized={isHeaderMinimized}
            onToggleHeader={() => setIsHeaderMinimized(!isHeaderMinimized)}
            categories={categories}
            taxonomies={taxonomies}
            measures={measures}
            onMeasuresChange={setMeasures}
            converters={converters}
            onConvertersChange={setConverters}
          />
        );
      }

      if (node.type === 'bucket_integration') {
        return (
          <IntegrationView
            node={node}
            breadcrumbs={bc}
            helpMsg={helpMsg}
            parentLabel={parentLabel}
            isHeaderMinimized={isHeaderMinimized}
            onToggleHeader={() => setIsHeaderMinimized(!isHeaderMinimized)}
          />
        );
      }

      return null;
    };

    if (selectedNode.type.startsWith('bucket_')) {
      return renderBucketView(selectedNode);
    }

    return (
      <div className="detail-pane flex flex-col h-full bg-[var(--surface)] p-10 overflow-hidden">
        <div className="detail-header flex flex-col gap-2.5 mb-10 shrink-0">
          <div className="breadcrumbs flex items-center gap-2 text-[11px] text-[var(--muted)] uppercase tracking-wider">
            {bc}
          </div>
          <h1 className="text-[32px] font-semibold tracking-tight m-0">{title}</h1>
        </div>
        <div className="flex flex-col items-center justify-center h-full text-[var(--muted)] gap-5 opacity-60">
          <div className="text-[64px]">{selectedNode.icon}</div>
          <div className="text-center">
            <h2 className="text-[24px] font-medium text-[var(--text)] mb-2">
              {selectedNode.label}
            </h2>
            <p className="text-[14px]">
              Select a category from the sidebar to begin planning or defining your environment.
            </p>
          </div>
        </div>
      </div>
    );
  };

  return (
    <div className="flex flex-col h-screen overflow-hidden">
      <Header />
      <main className="flex flex-1 overflow-hidden">
        <Sidebar
          tree={treeData}
          onSelect={handleNodeSelect}
          selectedNodeId={selectedNode?.id}
          isCollapsed={isSidebarCollapsed}
          onToggleCollapse={() => setIsSidebarCollapsed(!isSidebarCollapsed)}
        />
        <div
          id="detail-panel"
          className="flex-1 p-5 px-6 overflow-y-auto bg-[var(--bg)] scrollbar-thin scrollbar-thumb-[var(--border)]"
        >
          {renderDetailContent()}
        </div>
      </main>
      <Footer
        effectiveDate={effectiveDate}
        onDateChange={handleDateChange}
        helpMsg={
          selectedNode
            ? HELP_MESSAGES[selectedNode.type] || 'Select an item to explore.'
            : 'Click a node in the tree to explore.'
        }
      />

      <MetricModal
        isOpen={isMetricModalOpen}
        onClose={() => setIsMetricModalOpen(false)}
        onConfirm={handleMetricConfirm}
        onDelete={handleMetricDelete}
        existingMetric={editingMetric}
      />

      <CategoryModal
        isOpen={isCatModalOpen}
        onClose={() => setIsCatModalOpen(false)}
        onConfirm={handleCatConfirm}
        onDelete={handleCatDelete}
        mode={catModalMode}
        existingData={
          catModalMode.includes('Cat')
            ? categories.find(c => c.id === editingCatId)
            : categories
                .find(c => c.id === editingCatId)
                ?._children.find(el => el.id === editingElId)
        }
        initialTab={initialCatTab}
      />

      <TaxonomyModal
        isOpen={isTaxModalOpen}
        onClose={() => setIsTaxModalOpen(false)}
        onConfirm={handleTaxConfirm}
        onDelete={handleTaxDelete}
        existingTaxonomy={editingTaxonomy}
        categories={categories}
        initialTab={initialTaxTab}
      />

      <TaxonomyElementModal
        isOpen={isTaxElModalOpen}
        onClose={() => setIsTaxElModalOpen(false)}
        onConfirm={onTaxElConfirm}
        allowedCats={taxonomies.find(t => t.id === activeTaxId)?.allowedCats || []}
        categories={categories}
      />

      <PlanModal
        isOpen={isPlanModalOpen}
        onClose={() => setIsPlanModalOpen(false)}
        onConfirm={handlePlanConfirm}
        existingPlan={editingPlan}
      />
    </div>
  );
}
