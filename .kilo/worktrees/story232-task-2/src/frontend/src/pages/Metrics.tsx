import './Metrics.css';

export function Metrics() {
  return (
    <div className="metrics-page" data-testid="metrics-page">
      <header className="metrics-page__header">
        <h1 className="metrics-page__title">Metrics</h1>
      </header>

      <section className="metrics-page__content">
        <div className="metrics-page__placeholder">
          <div className="metrics-page__placeholder-icon">📊</div>
          <p className="metrics-page__placeholder-text">
            Metrics functionality coming some day soon.
          </p>
          <p className="metrics-page__placeholder-subtext">
            Stay tuned for powerful analytics and insights.
          </p>
        </div>
      </section>
    </div>
  );
}

export default Metrics;
