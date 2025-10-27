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
import globalMusicManager from "./services/globalMusicManager.ts";
import { isLandingPage } from "./services/musicConstants";
import "./App.css";

function App() {
  const [isWebSocketReady, setIsWebSocketReady] = useState(false);

  // Initialize audio settings, music manager, and WebSocket connection once on app startup
  useEffect(() => {
    // Initialize audio settings from localStorage
    audioSettingsManager.init();

    // Initialize global music manager
    globalMusicManager.init();

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

// Component that handles background visibility and music control based on route
function AppWithBackground() {
  const location = useLocation();

  // Show space background for landing pages
  const showSpaceBackground = isLandingPage(location.pathname);

  // Control music based on route changes and user interaction
  useEffect(() => {
    // Landing pages: start music (restart if coming from active game)
    if (isLandingPage(location.pathname)) {
      // Only start music on user interaction to comply with browser policies
      const handleInteraction = () => {
        const shouldRestart = globalMusicManager.shouldRestart();
        void globalMusicManager.startMusic(shouldRestart);
      };

      document.addEventListener("click", handleInteraction, { once: true });
      document.addEventListener("keydown", handleInteraction, { once: true });

      return () => {
        document.removeEventListener("click", handleInteraction);
        document.removeEventListener("keydown", handleInteraction);
      };
    }

    // Note: /game route music is handled by GameInterface component
    // because it needs to distinguish between lobby (music continues) and active (music stops)
  }, [location.pathname]);

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
