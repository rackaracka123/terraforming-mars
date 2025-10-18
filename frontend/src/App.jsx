import {
  BrowserRouter as Router,
  Routes,
  Route,
  useLocation,
} from "react-router-dom";
import { useEffect, useState } from "react";
import GameInterface from "./components/layout/main/GameInterface.tsx";
import CreateGamePage from "./components/pages/CreateGamePage.tsx";
import JoinGamePage from "./components/pages/JoinGamePage.tsx";
import CardsPage from "./components/pages/CardsPage.tsx";
import GameLandingPage from "./components/pages/GameLandingPage.tsx";
import ReconnectingPage from "./components/pages/ReconnectingPage.tsx";
import { globalWebSocketManager } from "./services/globalWebSocketManager.ts";
import { SpaceBackgroundProvider } from "./contexts/SpaceBackgroundContext.tsx";
import SpaceBackground from "./components/3d/SpaceBackground.tsx";
import audioSettingsManager from "./services/audioSettingsManager.ts";
import "./App.css";

function App() {
  const [isWebSocketReady, setIsWebSocketReady] = useState(false);

  // Initialize audio settings and WebSocket connection once on app startup
  useEffect(() => {
    // Initialize audio settings from localStorage
    audioSettingsManager.init();

    const initializeWebSocket = async () => {
      try {
        // console.log("Initializing global WebSocket connection...");
        await globalWebSocketManager.initialize();
        // console.log("Global WebSocket connection ready");
        setIsWebSocketReady(true);
      } catch (error) {
        console.error("Failed to initialize WebSocket:", error);
        // Continue running app even if WebSocket fails initially
        // It will retry connection when needed
        setIsWebSocketReady(true); // Allow app to continue
      }
    };

    void initializeWebSocket();
  }, []); // Empty dependency array - runs once on app mount

  // Show loading while WebSocket is initializing
  if (!isWebSocketReady) {
    return (
      <div
        style={{
          display: "flex",
          justifyContent: "center",
          alignItems: "center",
          height: "100vh",
          background: "#000011",
          color: "white",
          fontSize: "18px",
        }}
      >
        Connecting to server...
      </div>
    );
  }

  return (
    <SpaceBackgroundProvider>
      <div className="App" style={{ margin: 0, padding: 0 }}>
        <Router>
          <AppWithBackground />
        </Router>
      </div>
    </SpaceBackgroundProvider>
  );
}

// Component that handles background visibility based on route
function AppWithBackground() {
  const location = useLocation();

  // Show space background for landing, create, and join pages
  const showSpaceBackground = ["/", "/create", "/join"].includes(
    location.pathname,
  );

  return (
    <>
      {showSpaceBackground && (
        <SpaceBackground animationSpeed={0.5} overlayOpacity={0.3} />
      )}
      <Routes>
        <Route path="/" element={<GameLandingPage />} />
        <Route path="/create" element={<CreateGamePage />} />
        <Route path="/join" element={<JoinGamePage />} />
        <Route path="/cards" element={<CardsPage />} />
        <Route path="/reconnecting" element={<ReconnectingPage />} />
        <Route path="/game" element={<GameInterface />} />
      </Routes>
    </>
  );
}

export default App;
