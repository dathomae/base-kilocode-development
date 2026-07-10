import {
  ConnectionLineType,
  ReactFlow,
  ReactFlowProvider,
  useEdgesState,
  useNodesState,
  useReactFlow,
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import dagre from 'dagre';
import React, { useEffect, useMemo, useState } from 'react';
import { FLOW_EDGES, FLOW_NODES } from '../../../data/flowData';
import { cn } from '../../../lib/utils';

type TraceType = 'Tree' | 'Radial' | 'Sequence';

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const getLayoutedElements = (nodes: any[], edges: any[], direction = 'LR', type: TraceType) => {
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

const TraceContent: React.FC = () => {
  const [traceType, setTraceType] = useState<TraceType>('Tree');
  const [nodes, setNodes, onNodesChange] = useNodesState([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState([]);
  const { fitView } = useReactFlow();

  useEffect(() => {
    const direction = traceType === 'Sequence' ? 'LR' : 'TB';
    const { nodes: layoutedNodes, edges: layoutedEdges } = getLayoutedElements(
      FLOW_NODES,
      FLOW_EDGES,
      direction,
      traceType
    );
    setNodes(layoutedNodes);
    setEdges(layoutedEdges);

    setTimeout(() => fitView({ duration: 800 }), 50);
  }, [traceType, setNodes, setEdges, fitView]);

  const helpText = useMemo(() => {
    switch (traceType) {
      case 'Tree':
        return 'Best for tracing backward and debugging';
      case 'Radial':
        return 'Best for change impact tracing. What will happen if I change this?';
      case 'Sequence':
        return 'Best for understanding lifecycle';
      default:
        return '';
    }
  }, [traceType]);

  return (
    <div className="flex flex-col h-full relative p-4 bg-[var(--surface)]">
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center gap-1 bg-[var(--surface2)] p-1 rounded-lg border border-[var(--border)]">
          {(['Tree', 'Radial', 'Sequence'] as TraceType[]).map(type => (
            <button
              key={type}
              onClick={() => setTraceType(type)}
              className={cn(
                'px-3 py-1 rounded-md text-[11px] font-medium transition-all cursor-pointer',
                traceType === type
                  ? 'bg-[var(--accent2)] text-white shadow-sm'
                  : 'text-[var(--muted)] hover:text-[var(--text)] hover:bg-[var(--surface)]'
              )}
            >
              {type}
            </button>
          ))}
        </div>
        <div className="text-[11px] text-[var(--muted)] uppercase tracking-widest font-bold">
          Execution Trace
        </div>
      </div>

      <div className="flex-1 border border-[var(--border)] rounded-xl bg-[var(--bg)] overflow-hidden">
        <ReactFlow
          nodes={nodes}
          edges={edges}
          onNodesChange={onNodesChange}
          onEdgesChange={onEdgesChange}
          connectionLineType={ConnectionLineType.SmoothStep}
          colorMode="dark"
        />
      </div>

      <div className="mt-4 p-4 bg-[var(--surface2)] border border-[var(--border)] rounded-lg text-[13px] text-[var(--muted)] flex items-center gap-3 animate-in fade-in slide-in-from-bottom-2 duration-300">
        <div className="w-2 h-2 rounded-full bg-[var(--accent2)] animate-pulse" />
        <span className="font-medium text-[var(--text)] uppercase tracking-wider text-[10px] opacity-70">
          {traceType} Insight:
        </span>
        {helpText}
      </div>
    </div>
  );
};

export const PlanExecution: React.FC = () => (
  <ReactFlowProvider>
    <TraceContent />
  </ReactFlowProvider>
);
