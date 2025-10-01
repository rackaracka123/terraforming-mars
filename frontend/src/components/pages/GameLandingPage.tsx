import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { apiService } from "../../services/apiService";
import { globalWebSocketManager } from "../../services/globalWebSocketManager.ts";
import { useSpaceBackground } from "../../contexts/SpaceBackgroundContext.tsx";

const GameLandingPage: React.FC = () => {
  const navigate = useNavigate();
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isFadingOut, setIsFadingOut] = useState(false);
  const { preloadSkybox, isLoaded: isSkyboxLoaded } = useSpaceBackground();

  useEffect(() => {
    const checkExistingGame = async () => {
      try {
        // Preload skybox in parallel with game check
        void preloadSkybox();

        // Check localStorage for existing game
        const savedGameData = localStorage.getItem("terraforming-mars-game");
        if (savedGameData) {
          const { gameId, playerId, playerName } = JSON.parse(savedGameData);

          if (gameId && playerId) {
            // Try to get the current game state from server
            const game = await apiService.getGame(gameId);
            if (!game) {
              throw new Error("Saved game not found on server");
            }

            // Reconnect to the game using global WebSocket manager with playerId for proper reconnection
            await globalWebSocketManager.playerConnect(
              playerName,
              gameId,
              playerId,
            );

            // Navigate to game interface with the retrieved game state
            navigate("/game", {
              state: {
                game,
                playerId,
                playerName,
              },
            });
            return;
          }
        }

        setIsLoading(false);
      } catch (err: any) {
        void err;
        // Clear invalid saved game data
        localStorage.removeItem("terraforming-mars-game");
        setError("Unable to reconnect to previous game");
        setIsLoading(false);
      }
    };

    void checkExistingGame();
  }, [navigate, preloadSkybox]);

  const handleCreateGame = () => {
    setIsFadingOut(true);
    setTimeout(() => {
      navigate("/create");
    }, 300); // Match CSS transition duration
  };

  const handleJoinGame = () => {
    setIsFadingOut(true);
    setTimeout(() => {
      navigate("/join");
    }, 300); // Match CSS transition duration
  };

  const handleViewCards = () => {
    navigate("/cards");
  };

  if (isLoading) {
    return (
      <div
        style={{
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          minHeight: "100vh",
          color: "white",
        }}
      >
        <div style={{ textAlign: "center" }}>
          <h1
            style={{
              fontFamily: "'Orbitron', sans-serif",
              fontSize: "48px",
              marginBottom: "20px",
              fontWeight: 700,
              letterSpacing: "2px",
            }}
          >
            TERRAFORMING MARS
          </h1>
          <div
            style={{
              color: "#4a90e2",
              padding: "12px",
              background: "rgba(30, 60, 150, 0.1)",
              border: "1px solid rgba(30, 60, 150, 0.3)",
              borderRadius: "8px",
              fontSize: "16px",
            }}
          >
            Checking for existing game...
          </div>
        </div>
      </div>
    );
  }

  return (
      <div
        style={{
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          minHeight: "100vh",
          color: "white",
          fontFamily:
            "-apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', sans-serif",
          opacity: isFadingOut ? 0 : 1,
          transition: "opacity 0.3s ease-out",
          position: "relative",
          zIndex: 10,
        }}
      >
        <div style={{ textAlign: "center", maxWidth: "500px", padding: "20px" }}>
          {/* Title with Orbitron font */}
          <h1
            style={{
              fontFamily: "'Orbitron', sans-serif",
              fontSize: "56px",
              color: "#ffffff",
              marginBottom: "60px",
              textShadow: "0 0 30px rgba(30, 60, 150, 0.8)",
              fontWeight: 700,
              letterSpacing: "3px",
            }}
          >
            TERRAFORMING MARS
          </h1>

          {error && (
            <div
              style={{
                color: "#ff6b6b",
                marginBottom: "30px",
                padding: "12px",
                background: "rgba(255, 107, 107, 0.1)",
                border: "1px solid rgba(255, 107, 107, 0.3)",
                borderRadius: "8px",
                fontSize: "14px",
              }}
            >
              {error}
            </div>
          )}

          {/* Main action buttons - smaller and darker */}
          <div
            style={{
              display: "flex",
              gap: "20px",
              marginBottom: "40px",
              justifyContent: "center",
            }}
          >
            <button
              onClick={handleCreateGame}
              style={{
                background: "rgba(10, 10, 15, 0.9)",
                border: "2px solid rgba(30, 60, 150, 0.5)",
                borderRadius: "12px",
                padding: "20px 40px",
                cursor: "pointer",
                transition: "all 0.3s ease",
                backdropFilter: "blur(10px)",
                color: "white",
                fontSize: "18px",
                fontWeight: 600,
                fontFamily: "'Orbitron', sans-serif",
                letterSpacing: "1px",
              }}
              onMouseEnter={(e) => {
                e.currentTarget.style.borderColor = "rgba(30, 60, 150, 0.9)";
                e.currentTarget.style.boxShadow =
                  "0 0 30px rgba(30, 60, 150, 0.6), 0 0 60px rgba(30, 60, 150, 0.3)";
                e.currentTarget.style.transform = "translateY(-4px)";
              }}
              onMouseLeave={(e) => {
                e.currentTarget.style.borderColor = "rgba(30, 60, 150, 0.5)";
                e.currentTarget.style.boxShadow = "none";
                e.currentTarget.style.transform = "translateY(0)";
              }}
            >
              CREATE
            </button>

            <button
              onClick={handleJoinGame}
              style={{
                background: "rgba(10, 10, 15, 0.9)",
                border: "2px solid rgba(30, 60, 150, 0.5)",
                borderRadius: "12px",
                padding: "20px 40px",
                cursor: "pointer",
                transition: "all 0.3s ease",
                backdropFilter: "blur(10px)",
                color: "white",
                fontSize: "18px",
                fontWeight: 600,
                fontFamily: "'Orbitron', sans-serif",
                letterSpacing: "1px",
              }}
              onMouseEnter={(e) => {
                e.currentTarget.style.borderColor = "rgba(30, 60, 150, 0.9)";
                e.currentTarget.style.boxShadow =
                  "0 0 30px rgba(30, 60, 150, 0.6), 0 0 60px rgba(30, 60, 150, 0.3)";
                e.currentTarget.style.transform = "translateY(-4px)";
              }}
              onMouseLeave={(e) => {
                e.currentTarget.style.borderColor = "rgba(30, 60, 150, 0.5)";
                e.currentTarget.style.boxShadow = "none";
                e.currentTarget.style.transform = "translateY(0)";
              }}
            >
              JOIN
            </button>
          </div>
        </div>

        {/* View Cards button - bottom right corner */}
        <button
          onClick={handleViewCards}
          style={{
            position: "fixed",
            bottom: "30px",
            right: "30px",
            background: "rgba(10, 10, 15, 0.8)",
            border: "1px solid rgba(255, 255, 255, 0.2)",
            borderRadius: "8px",
            padding: "12px 20px",
            color: "rgba(255, 255, 255, 0.7)",
            cursor: "pointer",
            transition: "all 0.3s ease",
            fontSize: "14px",
            backdropFilter: "blur(5px)",
            fontFamily: "'Orbitron', sans-serif",
          }}
          onMouseEnter={(e) => {
            e.currentTarget.style.color = "rgba(255, 255, 255, 0.95)";
            e.currentTarget.style.borderColor = "rgba(30, 60, 150, 0.6)";
            e.currentTarget.style.background = "rgba(10, 10, 15, 0.95)";
            e.currentTarget.style.boxShadow =
              "0 0 20px rgba(30, 60, 150, 0.4)";
          }}
          onMouseLeave={(e) => {
            e.currentTarget.style.color = "rgba(255, 255, 255, 0.7)";
            e.currentTarget.style.borderColor = "rgba(255, 255, 255, 0.2)";
            e.currentTarget.style.background = "rgba(10, 10, 15, 0.8)";
            e.currentTarget.style.boxShadow = "none";
          }}
        >
          View Cards
        </button>
      </div>
  );
};

export default GameLandingPage;
