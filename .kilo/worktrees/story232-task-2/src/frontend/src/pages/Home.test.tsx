import { render, screen } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import { Home } from './Home';

describe('Home page', () => {
  it('renders the home page container', () => {
    render(<Home />);

    const homePage = screen.getByTestId('home-page');

    expect(homePage).toBeInTheDocument();
  });

  it('displays the welcome message title', () => {
    render(<Home />);

    const title = screen.getByRole('heading', { level: 1, name: /welcome to chronocone/i });

    expect(title).toBeInTheDocument();
  });

  it('displays the subtitle with framework description', () => {
    render(<Home />);

    const subtitle = screen.getByText(/a flexible framework for automated planning/i);

    expect(subtitle).toBeInTheDocument();
  });

  it('displays the About ChronoCone section', () => {
    render(<Home />);

    const aboutHeading = screen.getByRole('heading', { level: 2, name: /about chronocone/i });
    const aboutText = screen.getByText(/chronocone provides powerful tools for managing metrics/i);

    expect(aboutHeading).toBeInTheDocument();
    expect(aboutText).toBeInTheDocument();
  });

  it('displays the Metrics Management feature', () => {
    render(<Home />);

    const featureHeading = screen.getByRole('heading', { level: 3, name: /metrics management/i });
    const featureText = screen.getByText(/track and analyze your key performance indicators/i);

    expect(featureHeading).toBeInTheDocument();
    expect(featureText).toBeInTheDocument();
  });

  it('displays the Definitions feature', () => {
    render(<Home />);

    const featureHeading = screen.getByRole('heading', { level: 3, name: /definitions/i });
    const featureText = screen.getByText(/organize and manage your metadata types/i);

    expect(featureHeading).toBeInTheDocument();
    expect(featureText).toBeInTheDocument();
  });

  it('displays the Collaborative Planning feature', () => {
    render(<Home />);

    const featureHeading = screen.getByRole('heading', {
      level: 3,
      name: /collaborative planning/i,
    });
    const featureText = screen.getByText(/work together across teams/i);

    expect(featureHeading).toBeInTheDocument();
    expect(featureText).toBeInTheDocument();
  });

  it('has the correct CSS class on the container', () => {
    render(<Home />);

    const homePage = screen.getByTestId('home-page');

    expect(homePage).toHaveClass('home-page');
  });
});
