import './ErrorMessage.css';

export interface ErrorMessageProps {
  /**
   * The error message to display
   */
  message: string;

  /**
   * Callback function when the retry button is clicked
   */
  onRetry?: () => void;

  /**
   * Whether the retry operation is currently in progress
   * @default false
   */
  isRetrying?: boolean;

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
 * ErrorMessage Component
 *
 * Displays an error message with an optional retry button.
 * Used to show API errors and other failure states to users.
 */
export function ErrorMessage({
  message,
  onRetry,
  isRetrying = false,
  className = '',
  'data-testid': dataTestId = 'error-message',
}: ErrorMessageProps) {
  const combinedClassName = `error-message ${className}`.trim();

  return (
    <div className={combinedClassName} data-testid={dataTestId} role="alert">
      <div className="error-message__content">
        <div className="error-message__icon" aria-hidden="true">
          ⚠️
        </div>
        <p className="error-message__text">{message}</p>
      </div>
      {onRetry && (
        <button
          className="error-message__retry-btn"
          onClick={onRetry}
          disabled={isRetrying}
          data-testid="error-retry-button"
        >
          {isRetrying ? 'Retrying...' : 'Try Again'}
        </button>
      )}
    </div>
  );
}

export default ErrorMessage;
