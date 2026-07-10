import './App.css';
import { Routes, Route } from 'react-router-dom';
import { NavigationBar } from '@/components/NavigationBar';
import { Home, Definitions, Metrics } from '@/pages';

function App() {
  return (
    <>
      <NavigationBar />
      <main className="app-main">
        <Routes>
          <Route path="/" element={<Home />} />
          <Route path="/definitions" element={<Definitions />} />
          <Route path="/metrics" element={<Metrics />} />
        </Routes>
      </main>
    </>
  );
}

export default App;
