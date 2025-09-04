import { BrowserRouter as Router, Routes, Route } from "react-router-dom";
import GameInterface from "./components/layout/main/GameInterface.tsx";
import CreateGamePage from "./components/pages/CreateGamePage.tsx";
import JoinGamePage from "./components/pages/JoinGamePage.tsx";
import GameLandingPage from "./components/pages/GameLandingPage.tsx";
import "./App.css";

function App() {
  return (
    <div className="App" style={{ margin: 0, padding: 0 }}>
      <Router>
        <Routes>
          <Route path="/" element={<GameLandingPage />} />
          <Route path="/create" element={<CreateGamePage />} />
          <Route path="/join" element={<JoinGamePage />} />
          <Route path="/game" element={<GameInterface />} />
        </Routes>
      </Router>
    </div>
  );
}

export default App;
