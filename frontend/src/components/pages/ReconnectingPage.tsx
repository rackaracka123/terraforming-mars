import React, { useState, useEffect, useRef } from "react";
import { useNavigate } from "react-router-dom";
import { globalWebSocketManager } from "../../services/globalWebSocketManager.ts";
import { apiService } from "../../services/apiService.ts";
import styles from "./ReconnectingPage.module.css";

interface ReconnectingPageProps {
  // Add any props if needed
}

const ReconnectingPage: React.FC<ReconnectingPageProps> = () => {
  const navigate = useNavigate();
  const [error, setError] = useState<string | null>(null);
  const [isReconnecting, setIsReconnecting] = useState(true);
  const reconnectionAttempted = useRef(false);

  useEffect(() => {
    const attemptReconnection = async () => {
      // Prevent multiple reconnection attempts (React Strict Mode protection)
      if (reconnectionAttempted.current) {
        return;
      }
      reconnectionAttempted.current = true;

      try {
        setError(null);
        setIsReconnecting(true);

        // Check localStorage for saved game data
        const savedGameData = localStorage.getItem("terraforming-mars-game");
        if (!savedGameData) {
          throw new Error("No saved game data found");
        }

        const { gameId, playerName } = JSON.parse(savedGameData);
        if (!gameId || !playerName) {
          throw new Error("Invalid saved game data");
        }

        // Verify game still exists
        const game = await apiService.getGame(gameId);
        if (!game) {
          throw new Error("Game no longer exists");
        }

        // Ensure WebSocket is ready and attempt reconnection
        // Attempting to reconnect: playerName, gameId
        const reconnectionResult = await globalWebSocketManager.playerReconnect(
          playerName,
          gameId,
        );
        // Reconnection successful

        if (reconnectionResult.game) {
          // CRITICAL FIX: Set the current player ID in globalWebSocketManager
          // This ensures the GameInterface component knows which player this client represents
          globalWebSocketManager.setCurrentPlayerId(reconnectionResult.playerId);

          // Update localStorage with fresh data
          localStorage.setItem(
            "terraforming-mars-game",
            JSON.stringify({
              gameId: reconnectionResult.game.id,
              playerId: reconnectionResult.playerId,
              playerName: reconnectionResult.playerName,
              timestamp: Date.now(),
            }),
          );

          // Navigate to game with reconnected state
          navigate("/game", {
            state: {
              game: reconnectionResult.game,
              playerId: reconnectionResult.playerId,
              playerName: reconnectionResult.playerName,
              isReconnection: true,
            },
            replace: true,
          });
        }
      } catch (error: any) {
        console.error("Reconnection failed:", error);
        setError(error.message || "Reconnection failed");
      } finally {
        setIsReconnecting(false);
      }
    };

    attemptReconnection();
  }, [navigate]);

  const handleReturnToMenu = () => {
    localStorage.removeItem("terraforming-mars-game");
    navigate("/", { replace: true });
  };

  return (
    <div className={styles.reconnectingPage}>
      <div className={styles.container}>
        <div className={styles.content}>
          <h1 className={styles.title}>Terraforming Mars</h1>

          {isReconnecting ? (
            <div className={styles.reconnectingSection}>
              <div className={styles.spinner}></div>
              <h2 className={styles.reconnectingTitle}>
                Reconnecting to game...
              </h2>
              <p className={styles.reconnectingSubtext}>
                Please wait while we restore your connection
              </p>
            </div>
          ) : error ? (
            <div className={styles.errorSection}>
              <div className={styles.errorIcon}>⚠️</div>
              <h2 className={styles.errorTitle}>Reconnection Failed</h2>
              <p className={styles.errorMessage}>{error}</p>
              <button
                className={styles.returnButton}
                onClick={handleReturnToMenu}
              >
                Return to Main Menu
              </button>
            </div>
          ) : null}
        </div>
      </div>
    </div>
  );
};

export default ReconnectingPage;
