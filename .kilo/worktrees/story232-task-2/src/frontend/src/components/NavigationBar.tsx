import { Link } from 'react-router-dom';
import './NavigationBar.css';

export interface NavigationBarProps {
  className?: string;
}

export function NavigationBar({ className = '' }: NavigationBarProps) {
  return (
    <nav className={`navigation-bar ${className}`.trim()} data-testid="navigation-bar">
      <div className="navigation-bar__brand">
        <span className="navigation-bar__brand-text">ChronoCone</span>
      </div>
      <ul className="navigation-bar__links" role="menubar">
        <li role="none">
          <Link to="/" className="navigation-bar__link" role="menuitem" data-testid="nav-home">
            Home
          </Link>
        </li>
        <li role="none">
          <Link
            to="/definitions"
            className="navigation-bar__link"
            role="menuitem"
            data-testid="nav-definitions"
          >
            Definitions
          </Link>
        </li>
        <li role="none">
          <Link
            to="/metrics"
            className="navigation-bar__link"
            role="menuitem"
            data-testid="nav-metrics"
          >
            Metrics
          </Link>
        </li>
      </ul>
    </nav>
  );
}

export default NavigationBar;
