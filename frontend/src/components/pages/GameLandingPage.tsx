import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { apiService } from "../../services/apiService";
import { globalWebSocketManager } from "../../services/globalWebSocketManager.ts";

const GameLandingPage: React.FC = () => {
  const navigate = useNavigate();
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const checkExistingGame = async () => {
      try {
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

    checkExistingGame();
  }, [navigate]);

  const handleCreateGame = () => {
    navigate("/create");
  };

  const handleJoinGame = () => {
    navigate("/join");
  };

  const handleViewCards = () => {
    navigate("/cards");
  };

  if (isLoading) {
    return (
      <div className="game-landing-page">
        <div className="container">
          <div className="content">
            <h1>Terraforming Mars</h1>
            <div className="loading-message">Checking for existing game...</div>
          </div>
        </div>
        <style>{`
          .game-landing-page {
            background: #000011;
            color: white;
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            font-family:
              -apple-system, BlinkMacSystemFont, "Segoe UI", "Roboto", "Oxygen",
              "Ubuntu", "Cantarell", "Fira Sans", "Droid Sans",
              "Helvetica Neue", sans-serif;
          }

          .container {
            max-width: 600px;
            width: 100%;
            padding: 40px 20px;
          }

          .content {
            text-align: center;
          }

          h1 {
            font-size: 48px;
            color: #ffffff;
            margin-bottom: 40px;
            text-shadow: 0 2px 4px rgba(0, 0, 0, 0.8);
            font-weight: bold;
          }

          .loading-message {
            color: #4a90e2;
            padding: 12px;
            background: rgba(74, 144, 226, 0.1);
            border: 1px solid rgba(74, 144, 226, 0.3);
            border-radius: 8px;
            font-size: 16px;
            animation: pulse 2s infinite;
          }

          @keyframes pulse {
            0%,
            100% {
              opacity: 1;
            }
            50% {
              opacity: 0.7;
            }
          }
        `}</style>
      </div>
    );
  }

  return (
    <div className="game-landing-page">
      <div className="container">
        <div className="content">
          <h1>Terraforming Mars</h1>
          <p className="subtitle">
            Begin your journey to terraform the Red Planet
          </p>

          {error && <div className="error-message">{error}</div>}

          <div className="action-buttons">
            <div className="main-actions">
              <button
                onClick={handleCreateGame}
                className="action-button create-button"
              >
                <div className="button-content">
                  <h3>Create Game</h3>
                  <p>Start a new terraforming mission</p>
                </div>
              </button>

              <button
                onClick={handleJoinGame}
                className="action-button join-button"
              >
                <div className="button-content">
                  <h3>Join Game</h3>
                  <p>Join an existing mission</p>
                </div>
              </button>
            </div>

            <div className="secondary-actions">
              <button
                onClick={handleViewCards}
                className="secondary-button cards-button"
              >
                View Cards
              </button>
            </div>
          </div>
        </div>
      </div>

      <style>{`
        .game-landing-page {
          background: #000011;
          color: white;
          min-height: 100vh;
          display: flex;
          align-items: center;
          justify-content: center;
          font-family:
            -apple-system, BlinkMacSystemFont, "Segoe UI", "Roboto", "Oxygen",
            "Ubuntu", "Cantarell", "Fira Sans", "Droid Sans", "Helvetica Neue",
            sans-serif;
        }

        .container {
          max-width: 800px;
          width: 100%;
          padding: 40px 20px;
        }

        .content {
          text-align: center;
        }

        h1 {
          font-size: 64px;
          color: #ffffff;
          margin-bottom: 16px;
          text-shadow: 0 2px 4px rgba(0, 0, 0, 0.8);
          font-weight: bold;
        }

        .subtitle {
          font-size: 18px;
          color: rgba(255, 255, 255, 0.7);
          margin-bottom: 60px;
        }

        .error-message {
          color: #ff6b6b;
          margin-bottom: 40px;
          padding: 16px;
          background: rgba(255, 107, 107, 0.1);
          border: 1px solid rgba(255, 107, 107, 0.3);
          border-radius: 8px;
          font-size: 14px;
        }

        .action-buttons {
          display: flex;
          flex-direction: column;
          align-items: center;
          gap: 30px;
          max-width: 700px;
          margin: 0 auto;
        }

        .main-actions {
          display: grid;
          grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
          gap: 30px;
          width: 100%;
        }

        .secondary-actions {
          display: flex;
          justify-content: center;
          width: 100%;
        }

        .action-button {
          background: linear-gradient(
            135deg,
            rgba(20, 30, 50, 0.95) 0%,
            rgba(30, 40, 60, 0.93) 50%,
            rgba(25, 35, 55, 0.95) 100%
          );
          border: 2px solid rgba(100, 150, 255, 0.3);
          border-radius: 16px;
          padding: 40px 30px;
          cursor: pointer;
          transition: all 0.3s ease;
          backdrop-filter: blur(10px);
          color: white;
          text-decoration: none;
          display: block;
        }

        .action-button:hover {
          transform: translateY(-4px);
          border-color: rgba(100, 150, 255, 0.6);
          box-shadow:
            0 10px 30px rgba(0, 0, 0, 0.3),
            0 0 40px rgba(100, 150, 255, 0.2);
        }

        .create-button:hover {
          border-color: rgba(76, 175, 80, 0.6);
          box-shadow:
            0 10px 30px rgba(0, 0, 0, 0.3),
            0 0 40px rgba(76, 175, 80, 0.2);
        }

        .join-button:hover {
          border-color: rgba(255, 152, 0, 0.6);
          box-shadow:
            0 10px 30px rgba(0, 0, 0, 0.3),
            0 0 40px rgba(255, 152, 0, 0.2);
        }

        .secondary-button {
          background: transparent;
          border: 1px solid rgba(255, 255, 255, 0.2);
          border-radius: 8px;
          padding: 12px 24px;
          color: rgba(255, 255, 255, 0.7);
          cursor: pointer;
          transition: all 0.3s ease;
          font-size: 14px;
          backdrop-filter: blur(5px);
          text-decoration: none;
          display: inline-block;
        }

        .secondary-button:hover {
          color: rgba(255, 255, 255, 0.9);
          border-color: rgba(255, 255, 255, 0.4);
          background: rgba(255, 255, 255, 0.05);
          transform: translateY(-1px);
        }

        .button-content h3 {
          font-size: 24px;
          margin-bottom: 8px;
          font-weight: bold;
        }

        .button-content p {
          font-size: 14px;
          color: rgba(255, 255, 255, 0.7);
          margin: 0;
        }

        @media (max-width: 768px) {
          h1 {
            font-size: 48px;
          }

          .subtitle {
            font-size: 16px;
            margin-bottom: 40px;
          }

          .main-actions {
            grid-template-columns: 1fr;
            gap: 20px;
            max-width: 400px;
          }

          .action-button {
            padding: 30px 24px;
          }

          .button-content h3 {
            font-size: 20px;
          }

          .button-content p {
            font-size: 13px;
          }
        }

        @media (max-width: 480px) {
          .container {
            padding: 20px 15px;
          }

          h1 {
            font-size: 36px;
          }

          .subtitle {
            font-size: 14px;
          }

          .action-button {
            padding: 24px 20px;
          }

          .button-content h3 {
            font-size: 18px;
          }

          .button-content p {
            font-size: 12px;
          }
        }
      `}</style>
    </div>
  );
};

export default GameLandingPage;
