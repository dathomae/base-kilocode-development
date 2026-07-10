import { render, screen } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import { Metrics } from './Metrics';

describe('Metrics page', () => {
  it('renders the metrics page container', () => {
    render(<Metrics />);

    const metricsPage = screen.getByTestId('metrics-page');

    expect(metricsPage).toBeInTheDocument();
  });

  it('displays the page title', () => {
    render(<Metrics />);

    const title = screen.getByRole('heading', { level: 1, name: /metrics/i });

    expect(title).toBeInTheDocument();
  });

  it('displays the placeholder message', () => {
    render(<Metrics />);

    const placeholderText = screen.getByText(/metrics functionality coming some day soon/i);

    expect(placeholderText).toBeInTheDocument();
  });

  it('displays the subtext about analytics', () => {
    render(<Metrics />);

    const subtext = screen.getByText(/stay tuned for powerful analytics and insights/i);

    expect(subtext).toBeInTheDocument();
  });

  it('has the correct CSS class on the container', () => {
    render(<Metrics />);

    const metricsPage = screen.getByTestId('metrics-page');

    expect(metricsPage).toHaveClass('metrics-page');
  });
});
