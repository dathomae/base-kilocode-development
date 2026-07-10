import './Home.css';

export function Home() {
  return (
    <div className="home-page" data-testid="home-page">
      <header className="home-page__header">
        <h1 className="home-page__title">Welcome to ChronoCone</h1>
        <p className="home-page__subtitle">
          A flexible framework for automated planning, human-in-the-loop planning, and collaborative
          planning across organizational levels.
        </p>
      </header>

      <section className="home-page__content">
        <div className="home-page__card">
          <h2 className="home-page__card-title">About ChronoCone</h2>
          <p className="home-page__card-text">
            ChronoCone provides powerful tools for managing metrics, definitions, and planning
            workflows. Whether you need automated data processing or collaborative planning
            features, ChronoCone adapts to your needs.
          </p>
        </div>

        <div className="home-page__features">
          <div className="home-page__feature">
            <h3 className="home-page__feature-title">Metrics Management</h3>
            <p className="home-page__feature-text">
              Track and analyze your key performance indicators with ease.
            </p>
          </div>
          <div className="home-page__feature">
            <h3 className="home-page__feature-title">Definitions</h3>
            <p className="home-page__feature-text">
              Organize and manage your metadata types and taxonomy structures.
            </p>
          </div>
          <div className="home-page__feature">
            <h3 className="home-page__feature-title">Collaborative Planning</h3>
            <p className="home-page__feature-text">
              Work together across teams with shared planning tools.
            </p>
          </div>
        </div>
      </section>
    </div>
  );
}

export default Home;
