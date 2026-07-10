import './LoadingSpinner.css';

export interface LoadingSpinnerProps {
  /**
   * Text to display below the spinner
   * @default "Loading..."
   */
  text?: string;

  /**
   * Size of the spinner
   * @default "medium"
   */
  size?: 'small' | 'medium' | 'large';

  /**
   * Whether to show the spinner as an overlay
   * @default false
   */
  overlay?: boolean;

  /**
   * Additional CSS class names
   */
  className?: string;

  /**
   * Test ID for testing
   */
  'data-testid'?: string;
}

/**
 * LoadingSpinner Component
 *
 * Displays a spinning loading indicator with optional text.
 * Used to indicate that data is being loaded or an operation is in progress.
 */
export function LoadingSpinner({
  text = 'Loading...',
  size = 'medium',
  overlay = false,
  className = '',
  'data-testid': dataTestId = 'loading-spinner',
}: LoadingSpinnerProps) {
  const sizeClass = `loading-spinner--${size}`;
  const overlayClass = overlay ? 'loading-spinner--overlay' : '';
  const combinedClassName = `loading-spinner ${sizeClass} ${overlayClass} ${className}`.trim();

  return (
    <div className={combinedClassName} data-testid={dataTestId}>
      <div className="loading-spinner__spinner" aria-hidden="true" />
      {text && <p className="loading-spinner__text">{text}</p>}
    </div>
  );
}

export default LoadingSpinner;
