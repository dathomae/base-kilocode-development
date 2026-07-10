import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { describe, it, expect } from 'vitest';
import { NavigationBar } from './NavigationBar';

describe('NavigationBar', () => {
  it('renders the navigation bar', () => {
    render(
      <MemoryRouter>
        <NavigationBar />
      </MemoryRouter>
    );

    expect(screen.getByTestId('navigation-bar')).toBeInTheDocument();
  });

  it('displays the brand name', () => {
    render(
      <MemoryRouter>
        <NavigationBar />
      </MemoryRouter>
    );

    expect(screen.getByText('ChronoCone')).toBeInTheDocument();
  });

  it('renders Home navigation link', () => {
    render(
      <MemoryRouter>
        <NavigationBar />
      </MemoryRouter>
    );

    const homeLink = screen.getByTestId('nav-home');
    expect(homeLink).toBeInTheDocument();
    expect(homeLink).toHaveTextContent('Home');
    expect(homeLink).toHaveAttribute('href', '/');
  });

  it('renders Definitions navigation link', () => {
    render(
      <MemoryRouter>
        <NavigationBar />
      </MemoryRouter>
    );

    const definitionsLink = screen.getByTestId('nav-definitions');
    expect(definitionsLink).toBeInTheDocument();
    expect(definitionsLink).toHaveTextContent('Definitions');
    expect(definitionsLink).toHaveAttribute('href', '/definitions');
  });

  it('renders Metrics navigation link', () => {
    render(
      <MemoryRouter>
        <NavigationBar />
      </MemoryRouter>
    );

    const metricsLink = screen.getByTestId('nav-metrics');
    expect(metricsLink).toBeInTheDocument();
    expect(metricsLink).toHaveTextContent('Metrics');
    expect(metricsLink).toHaveAttribute('href', '/metrics');
  });

  it('applies custom className when provided', () => {
    render(
      <MemoryRouter>
        <NavigationBar className="custom-class" />
      </MemoryRouter>
    );

    const navBar = screen.getByTestId('navigation-bar');
    expect(navBar).toHaveClass('navigation-bar');
    expect(navBar).toHaveClass('custom-class');
  });

  it('renders all navigation links in a list', () => {
    render(
      <MemoryRouter>
        <NavigationBar />
      </MemoryRouter>
    );

    const linksList = screen.getByRole('menubar');
    expect(linksList).toBeInTheDocument();

    const menuItems = screen.getAllByRole('menuitem');
    expect(menuItems).toHaveLength(3);
  });
});
