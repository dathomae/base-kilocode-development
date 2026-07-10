import { useState, useEffect, useCallback } from 'react';
import { MetadataTree, type TreeNode } from '../components/MetadataTree';
import { LoadingSpinner } from '../components/LoadingSpinner';
import { ErrorMessage } from '../components/ErrorMessage';
import {
  fetchMetadataHierarchy,
  createMetadata,
  toBackendType,
  isApiError,
  type HierarchyItem,
} from '../services/metricsApi';
import './Definitions.css';

/**
 * Build tree structure from flat hierarchy items
 * Uses names instead of IDs for parent-child relationships
 */
function buildTreeFromHierarchy(items: HierarchyItem[]): TreeNode[] {
  // Create maps for quick lookup using names as keys
  const nsFamMap = new Map<string, TreeNode>();
  const nsMap = new Map<string, TreeNode>();
  const categoryMap = new Map<string, TreeNode>();
  const elementMap = new Map<string, TreeNode>();

  // First pass: Create nodes for each type using names as identifiers
  items.forEach(item => {
    switch (item.type) {
      case 'namespace_family':
        if (item.ns_fam_name && !nsFamMap.has(item.ns_fam_name)) {
          nsFamMap.set(item.ns_fam_name, {
            type: 'namespace_family',
            name: item.ns_fam_name,
            description: item.description,
            tooltip: item.tooltip,
            children: [],
            ns_fam_name: item.ns_fam_name,
          });
        }
        break;

      case 'namespace':
        if (item.ns_name && !nsMap.has(item.ns_name)) {
          nsMap.set(item.ns_name, {
            type: 'namespace',
            name: item.ns_name,
            description: item.description,
            tooltip: item.tooltip,
            children: [],
            ns_fam_name: item.ns_fam_name,
            ns_name: item.ns_name,
          });
        }
        break;

      case 'category':
        if (item.category_name && !categoryMap.has(item.category_name)) {
          categoryMap.set(item.category_name, {
            type: 'category',
            name: item.category_name,
            description: item.description,
            tooltip: item.tooltip,
            children: [],
            ns_fam_name: item.ns_fam_name,
            ns_name: item.ns_name,
            category_name: item.category_name,
          });
        }
        break;

      case 'element':
        if (item.element_name && !elementMap.has(item.element_name)) {
          elementMap.set(item.element_name, {
            type: 'element',
            name: item.element_name,
            description: item.description,
            tooltip: item.tooltip,
            children: [],
            ns_fam_name: item.ns_fam_name,
            ns_name: item.ns_name,
            category_name: item.category_name,
            element_name: item.element_name,
          });
        }
        break;
    }
  });

  // Second pass: Build hierarchy using parent names
  // Add elements to their parent categories
  elementMap.forEach(elementNode => {
    const parentCategoryName = elementNode.category_name;
    if (parentCategoryName && categoryMap.has(parentCategoryName)) {
      const categoryNode = categoryMap.get(parentCategoryName)!;
      categoryNode.children.push(elementNode);
    }
  });

  // Add categories to their parent namespaces
  categoryMap.forEach(categoryNode => {
    const parentNamespaceName = categoryNode.ns_name;
    if (parentNamespaceName && nsMap.has(parentNamespaceName)) {
      const nsNode = nsMap.get(parentNamespaceName)!;
      nsNode.children.push(categoryNode);
    }
  });

  // Add namespaces to their parent namespace families
  nsMap.forEach(nsNode => {
    const parentFamName = nsNode.ns_fam_name;
    if (parentFamName && nsFamMap.has(parentFamName)) {
      const nsFamNode = nsFamMap.get(parentFamName)!;
      nsFamNode.children.push(nsNode);
    }
  });

  // Return root nodes (namespace families)
  return Array.from(nsFamMap.values());
}

export function Definitions() {
  const [treeNodes, setTreeNodes] = useState<TreeNode[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isSaving, setIsSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  /**
   * Load metadata hierarchy from the API
   */
  const loadMetadata = useCallback(async (signal?: AbortSignal) => {
    setIsLoading(true);
    setError(null);

    try {
      const hierarchyItems = await fetchMetadataHierarchy();
      // Check if the component is still mounted before updating state
      if (signal?.aborted) return;
      const tree = buildTreeFromHierarchy(hierarchyItems);
      setTreeNodes(tree);
    } catch (err) {
      // Don't update state if aborted or unmounted
      if (signal?.aborted) return;
      if (isApiError(err)) {
        setError(
          err.isNetworkError
            ? 'Unable to connect to the server. Please check that the Metrics Service is running on port 8001.'
            : `Failed to load metadata: ${err.message}`
        );
      } else {
        setError('An unexpected error occurred while loading metadata.');
      }
    } finally {
      // Only update loading state if not aborted
      if (!signal?.aborted) {
        setIsLoading(false);
      }
    }
  }, []);

  // Load metadata on component mount with cleanup
  useEffect(() => {
    const abortController = new AbortController();
    loadMetadata(abortController.signal);
    return () => {
      abortController.abort();
    };
  }, [loadMetadata]);

  /**
   * Handle adding a new child item to a parent node
   */
  const handleAddChild = useCallback(
    async (
      parentType: string,
      parentNode: TreeNode | null,
      name: string,
      description: string,
      tooltip: string
    ) => {
      // Determine child type and required attributes based on parent
      let childType: string;
      let attributes: Record<string, unknown> = {};

      switch (parentType) {
        case 'namespace_family':
          if (parentNode === null) {
            // Creating a root namespace family
            childType = 'namespace_family';
            // No extra attributes needed for namespace_family
            attributes = {};
          } else {
            // Creating a namespace under a namespace family
            childType = 'namespace';
            // Namespace needs namespace_family_name
            attributes = {
              namespace_family_name: parentNode.name,
            };
          }
          break;

        case 'namespace':
          childType = 'category';
          // Category needs namespace_name
          attributes = {
            namespace_name: parentNode?.name ?? '',
          };
          break;

        case 'category':
          childType = 'element';
          // Element needs category_name and namespace_name
          attributes = {
            category_name: parentNode?.name ?? '',
            namespace_name: parentNode?.ns_name ?? '',
          };
          break;

        default:
          setError(`Cannot create children for type "${parentType}"`);
          return;
      }

      const backendType = toBackendType(childType);
      if (!backendType) {
        setError(`Cannot create items for type "${childType}" - not supported by backend`);
        return;
      }

      setIsSaving(true);
      setError(null);

      try {
        await createMetadata(backendType, name, description, tooltip, attributes);

        // Refresh the metadata tree to show the new item
        await loadMetadata();
      } catch (err) {
        if (isApiError(err)) {
          setError(
            err.isNetworkError
              ? 'Unable to connect to the server. Please check that the Metrics Service is running.'
              : `Failed to create metadata: ${err.message}`
          );
        } else {
          setError('An unexpected error occurred while saving.');
        }
      } finally {
        setIsSaving(false);
      }
    },
    [loadMetadata]
  );

  /**
   * Handle retry button click
   */
  const handleRetry = useCallback(() => {
    loadMetadata();
  }, [loadMetadata]);

  // Show loading spinner while fetching initial data
  if (isLoading) {
    return (
      <div className="definitions-page" data-testid="definitions-page">
        <header className="definitions-page__header">
          <h1 className="definitions-page__title">Definitions</h1>
          <p className="definitions-page__subtitle">
            Manage your metadata types, taxonomy structures, and organizational definitions.
          </p>
        </header>
        <section className="definitions-page__content">
          <LoadingSpinner text="Loading metadata..." />
        </section>
      </div>
    );
  }

  return (
    <div className="definitions-page" data-testid="definitions-page">
      <header className="definitions-page__header">
        <h1 className="definitions-page__title">Definitions</h1>
        <p className="definitions-page__subtitle">
          Manage your metadata types, taxonomy structures, and organizational definitions.
        </p>
      </header>

      <section className="definitions-page__content">
        {error && (
          <ErrorMessage
            message={error}
            onRetry={handleRetry}
            isRetrying={isLoading}
            className="definitions-page__error"
          />
        )}

        <div className="definitions-page__intro">
          <p className="definitions-page__intro-text">
            Browse the hierarchy below to view and manage metadata. Click the + button next to any
            item to add children. The structure shows namespace families, namespaces, categories,
            and elements in their hierarchical relationships.
          </p>
        </div>

        <div className="definitions-page__tree">
          <MetadataTree nodes={treeNodes} onAddChild={handleAddChild} isSaving={isSaving} />
        </div>
      </section>
    </div>
  );
}

export default Definitions;
