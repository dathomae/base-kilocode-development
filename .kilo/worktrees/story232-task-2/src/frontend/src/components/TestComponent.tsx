// Simple component to verify TypeScript and path aliases are working
import { useState } from 'react';

interface TestComponentProps {
  initialCount?: number;
}

export function TestComponent({ initialCount = 0 }: TestComponentProps) {
  const [count, setCount] = useState(initialCount);

  return (
    <div className="test-component" data-testid="test-component">
      <h2>Test Component Works!</h2>
      <p>Count: {count}</p>
      <button onClick={() => setCount(c => c + 1)}>Increment</button>
    </div>
  );
}

export default TestComponent;
