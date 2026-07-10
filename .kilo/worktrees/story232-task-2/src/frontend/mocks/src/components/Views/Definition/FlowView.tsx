import {
  ConnectionLineType,
  EdgeProps,
  Handle,
  Position,
  ReactFlow,
  ReactFlowProvider,
  getBezierPath,
  useEdgesState,
  useNodesState,
  useReactFlow,
} from '@xyflow/react';
import React, { useEffect, useMemo } from 'react';

import '@xyflow/react/dist/style.css';
import { sankey, sankeyJustify } from 'd3-sankey';
import dagre from 'dagre';

import { FLOW_EDGES, FLOW_NODES } from '../../../data/flowData';

export type FlowType = 'Sankey' | 'Tree' | 'Radial' | 'Sequence';

interface FlowViewProps {
  type: FlowType;
}

// Custom Sankey Node
const SankeyNode = ({ data }: { data: { label: string } }) => (
  <div className="flex flex-col items-center justify-center h-full w-full bg-[var(--accent2)] border border-[var(--accent2)] rounded-sm shadow-lg">
    <Handle type="target" position={Position.Left} className="opacity-0" />
    <div className="absolute -top-6 left-1/2 -translate-x-1/2 whitespace-nowrap text-[10px] font-bold text-[var(--text)] uppercase tracking-tighter bg-[var(--surface2)] px-1.5 py-0.5 rounded border border-[var(--border)]">
      {data.label}
    </div>
    <Handle type="source" position={Position.Right} className="opacity-0" />
  </div>
);

// Custom Sankey Edge
const SankeyEdge = ({
  id,
  sourceX,
  sourceY,
  targetX,
  targetY,
  sourcePosition,
  targetPosition,
  style = {},
  markerEnd,
  data,
}: EdgeProps) => {
  const [edgePath] = getBezierPath({
    sourceX,
    sourceY,
    targetX,
    targetY,
    sourcePosition,
    targetPosition,
  });

  const strokeWidth = Math.max(2, Number(data?.value || 1) * 2);

  return (
    <path
      id={id}
      style={{
        ...style,
        strokeWidth,
        stroke: 'var(--accent2)',
        strokeOpacity: 0.2,
        fill: 'none',
      }}
      className="react-flow__edge-path transition-all hover:stroke-opacity-50"
      d={edgePath}
      markerEnd={markerEnd}
    />
  );
};

const nodeTypes = {
  sankey: SankeyNode,
};

const edgeTypes = {
  sankey: SankeyEdge,
};

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const getLayoutedElements = (nodes: any[], edges: any[], direction = 'LR', type: FlowType) => {
  if (type === 'Sankey') {
    const width = 1000;
    const height = 600;

    const sankeyGenerator = sankey()
      .nodeWidth(12)
      .nodePadding(40)
      .extent([
        [0, 0],
        [width, height],
      ])
      .nodeAlign(sankeyJustify);

    const sankeyData = {
      nodes: nodes.map((n, i) => ({ id: n.id, index: i, name: n.data.label })),
      links: edges.map(e => ({
        source: nodes.findIndex(n => n.id === e.source),
        target: nodes.findIndex(n => n.id === e.target),
        value: e.data?.value || 1,
      })),
    };

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const { nodes: sNodes, links: sLinks } = sankeyGenerator(sankeyData as any);

    const layoutedNodes = (
      sNodes as { id: string; name: string; x0: number; y0: number; x1: number; y1: number }[]
    ).map(node => ({
      id: node.id,
      type: 'sankey',
      data: { label: node.name },
      position: { x: node.x0, y: node.y0 },
      style: { width: node.x1 - node.x0, height: node.y1 - node.y0 },
    }));

    const layoutedEdges = (
      sLinks as { source: { id: string }; target: { id: string }; value: number }[]
    ).map((link, i) => ({
      id: `e-${i}`,
      source: link.source.id,
      target: link.target.id,
      type: 'sankey',
      data: { value: link.value },
    }));

    return { nodes: layoutedNodes, edges: layoutedEdges };
  }

  const dagreGraph = new dagre.graphlib.Graph();
  dagreGraph.setDefaultEdgeLabel(() => ({}));

  const nodeWidth = 172;
  const nodeHeight = 36;

  if (type === 'Radial') {
    const radius = 400;
    const centerX = 500;
    const centerY = 500;
    return {
      nodes: nodes.map((node, index) => {
        const angle = (index / nodes.length) * 2 * Math.PI;
        return {
          ...node,
          position: {
            x: centerX + radius * Math.cos(angle),
            y: centerY + radius * Math.sin(angle),
          },
        };
      }),
      edges,
    };
  }

  dagreGraph.setGraph({ rankdir: direction, ranksep: 100, nodesep: 50 });

  nodes.forEach(node => {
    dagreGraph.setNode(node.id, { width: nodeWidth, height: nodeHeight });
  });

  edges.forEach(edge => {
    dagreGraph.setEdge(edge.source, edge.target);
  });

  dagre.layout(dagreGraph);

  const layoutedNodes = nodes.map(node => {
    const nodeWithPosition = dagreGraph.node(node.id);
    return {
      ...node,
      position: {
        x: nodeWithPosition.x - nodeWidth / 2,
        y: nodeWithPosition.y - nodeHeight / 2,
      },
    };
  });

  return { nodes: layoutedNodes, edges };
};

const FlowContent: React.FC<FlowViewProps> = ({ type }) => {
  const [nodes, setNodes, onNodesChange] = useNodesState([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState([]);
  const { fitView } = useReactFlow();

  useEffect(() => {
    const direction = type === 'Sequence' ? 'LR' : 'TB';
    const { nodes: layoutedNodes, edges: layoutedEdges } = getLayoutedElements(
      FLOW_NODES,
      FLOW_EDGES,
      direction,
      type
    );
    setNodes(layoutedNodes);
    setEdges(layoutedEdges);

    setTimeout(() => fitView({ duration: 800 }), 50);
  }, [type, setNodes, setEdges, fitView]);

  const helpText = useMemo(() => {
    switch (type) {
      case 'Sankey':
        return 'Best for understanding magnitude of downstream change';
      case 'Tree':
        return 'Best for tracing backward and debugging';
      case 'Radial':
        return 'Best for change impact tracing. What will happen if I change this?';
      case 'Sequence':
        return 'Best for understanding lifecycle';
      default:
        return '';
    }
  }, [type]);

  return (
    <div className="flex flex-col h-full relative">
      <div className="flex-1 border border-[var(--border)] rounded-xl bg-[var(--bg)] overflow-hidden">
        <ReactFlow
          nodes={nodes}
          edges={edges}
          onNodesChange={onNodesChange}
          onEdgesChange={onEdgesChange}
          nodeTypes={nodeTypes}
          edgeTypes={edgeTypes}
          connectionLineType={ConnectionLineType.SmoothStep}
          colorMode="dark"
        />
      </div>
      <div className="mt-4 p-4 bg-[var(--surface2)] border border-[var(--border)] rounded-lg text-[13px] text-[var(--muted)] flex items-center gap-3 animate-in fade-in slide-in-from-bottom-2 duration-300">
        <div className="w-2 h-2 rounded-full bg-[var(--accent2)] animate-pulse" />
        <span className="font-medium text-[var(--text)] uppercase tracking-wider text-[10px] opacity-70">
          {type} Insight:
        </span>
        {helpText}
      </div>
    </div>
  );
};

export const FlowView: React.FC<FlowViewProps> = props => (
  <ReactFlowProvider>
    <FlowContent {...props} />
  </ReactFlowProvider>
);
