import { Node, Edge } from '@xyflow/react';

export const FLOW_NODES: Node[] = Array.from({ length: 20 }, (_, i) => ({
  id: `node-${i + 1}`,
  data: { label: `Process ${i + 1}` },
  position: { x: 0, y: 0 },
  type: 'default',
}));

export const FLOW_EDGES: Edge[] = [
  { id: 'e1-2', source: 'node-1', target: 'node-2', data: { value: 10 } },
  { id: 'e1-3', source: 'node-1', target: 'node-3', data: { value: 15 } },
  { id: 'e2-4', source: 'node-2', target: 'node-4', data: { value: 5 } },
  { id: 'e2-5', source: 'node-2', target: 'node-5', data: { value: 5 } },
  { id: 'e3-6', source: 'node-3', target: 'node-6', data: { value: 8 } },
  { id: 'e3-7', source: 'node-3', target: 'node-7', data: { value: 7 } },
  { id: 'e4-8', source: 'node-4', target: 'node-8', data: { value: 3 } },
  { id: 'e5-9', source: 'node-5', target: 'node-9', data: { value: 4 } },
  { id: 'e6-10', source: 'node-6', target: 'node-10', data: { value: 6 } },
  { id: 'e7-11', source: 'node-7', target: 'node-11', data: { value: 5 } },
  { id: 'e8-12', source: 'node-8', target: 'node-12', data: { value: 2 } },
  { id: 'e9-13', source: 'node-9', target: 'node-13', data: { value: 3 } },
  { id: 'e10-14', source: 'node-10', target: 'node-14', data: { value: 4 } },
  { id: 'e11-15', source: 'node-11', target: 'node-15', data: { value: 3 } },
  { id: 'e12-16', source: 'node-12', target: 'node-16', data: { value: 2 } },
  { id: 'e13-17', source: 'node-13', target: 'node-17', data: { value: 2 } },
  { id: 'e14-18', source: 'node-14', target: 'node-18', data: { value: 3 } },
  { id: 'e15-19', source: 'node-15', target: 'node-19', data: { value: 2 } },
  { id: 'e16-20', source: 'node-16', target: 'node-20', data: { value: 1 } },
  { id: 'e1-4', source: 'node-1', target: 'node-4', data: { value: 5 } },
];
