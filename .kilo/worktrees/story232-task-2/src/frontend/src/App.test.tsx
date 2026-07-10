import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { describe, it, expect } from 'vitest';
import App from './App';

describe('App component', () => {
  it('renders the NavigationBar', () => {
    render(
      <MemoryRouter>
        <App />
      </MemoryRouter>
    );

    const navigationBar = screen.getByTestId('navigation-bar');

    expect(navigationBar).toBeInTheDocument();
  });

  it('renders Home page by default on root route', () => {
    render(
      <MemoryRouter initialEntries={['/']}>
        <App />
      </MemoryRouter>
    );

    const homePage = screen.getByTestId('home-page');

    expect(homePage).toBeInTheDocument();
    expect(screen.getByText('Welcome to ChronoCone')).toBeInTheDocument();
  });

  it('renders Definitions page on /definitions route', () => {
    render(
      <MemoryRouter initialEntries={['/definitions']}>
        <App />
      </MemoryRouter>
    );

    const definitionsPage = screen.getByTestId('definitions-page');

    expect(definitionsPage).toBeInTheDocument();
    // Check for specific text that only appears in Definitions page content
    expect(screen.getByText(/manage your metadata types/i)).toBeInTheDocument();
  });

  it('renders Metrics page on /metrics route', () => {
    render(
      <MemoryRouter initialEntries={['/metrics']}>
        <App />
      </MemoryRouter>
    );

    const metricsPage = screen.getByTestId('metrics-page');

    expect(metricsPage).toBeInTheDocument();
    // Check for specific text that only appears in Metrics page content
    expect(screen.getByText('Metrics functionality coming some day soon.')).toBeInTheDocument();
  });

  it('only shows one page at a time based on route', () => {
    const { unmount } = render(
      <MemoryRouter initialEntries={['/']}>
        <App />
      </MemoryRouter>
    );

    // Home page should be visible
    expect(screen.getByTestId('home-page')).toBeInTheDocument();
    expect(screen.queryByTestId('definitions-page')).not.toBeInTheDocument();
    expect(screen.queryByTestId('metrics-page')).not.toBeInTheDocument();

    // Unmount and render Definitions page
    unmount();
    render(
      <MemoryRouter initialEntries={['/definitions']}>
        <App />
      </MemoryRouter>
    );

    // Definitions page should be visible
    expect(screen.queryByTestId('home-page')).not.toBeInTheDocument();
    expect(screen.getByTestId('definitions-page')).toBeInTheDocument();
    expect(screen.queryByTestId('metrics-page')).not.toBeInTheDocument();
  });
});
