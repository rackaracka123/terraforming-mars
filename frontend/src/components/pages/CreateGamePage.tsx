import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import { apiService } from "../../services/apiService";
import { webSocketService } from "../../services/webSocketService";
import { GameSettings } from "../../types/generated/domain";

const CreateGamePage: React.FC = () => {
  const navigate = useNavigate();
  const [playerName, setPlayerName] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!playerName.trim()) {
      setError("Please enter your name");
      return;
    }

    if (playerName.trim().length < 2) {
      setError("Name must be at least 2 characters long");
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      // Create game with default settings
      const gameSettings: GameSettings = {
        maxPlayers: 4, // Default max players
      };

      // Creating game with settings
      const game = await apiService.createGame(gameSettings);
      // Game created successfully

      // Connect to WebSocket if not already connected
      if (!webSocketService.connected) {
        await webSocketService.connect();
      }

      // Connect player to the game
      const playerConnectedResult = await webSocketService.playerConnect(
        playerName.trim(),
        game.id,
      );
      // Player connected successfully

      // Save game data to localStorage for reconnection
      const gameData = {
        gameId: game.id,
        playerId: playerConnectedResult.playerId,
        playerName: playerName.trim(),
        createdAt: new Date().toISOString(),
      };
      localStorage.setItem("terraforming-mars-game", JSON.stringify(gameData));

      // Navigate to the main game interface
      navigate("/game", {
        state: {
          game,
          playerId: playerConnectedResult.playerId,
          playerName: playerName.trim(),
        },
      });
    } catch (err) {
      console.error("Failed to create game:", err);
      setError(err instanceof Error ? err.message : "Failed to create game");
    } finally {
      setIsLoading(false);
    }
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setPlayerName(e.target.value);
    if (error) setError(null); // Clear error when user starts typing
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Enter") {
      handleSubmit(e as React.FormEvent);
    }
  };

  return (
    <div className="create-game-page">
      <div className="container">
        <div className="content">
          <h1>Create a new game</h1>

          <form onSubmit={handleSubmit} className="create-game-form">
            <div className="input-container">
              <input
                type="text"
                value={playerName}
                onChange={handleInputChange}
                onKeyDown={handleKeyDown}
                placeholder="Enter your name"
                disabled={isLoading}
                className="player-name-input"
                autoFocus
                maxLength={50}
              />

              <button
                type="submit"
                disabled={isLoading || !playerName.trim()}
                className="submit-button"
                title="Connect"
              >
                <img
                  src="/assets/misc/arrow.png"
                  alt="Connect"
                  className="arrow-icon"
                />
              </button>
            </div>

            {error && <div className="error-message">{error}</div>}

            {isLoading && (
              <div className="loading-message">Creating game...</div>
            )}
          </form>
        </div>
      </div>

      <style jsx>{`
        .create-game-page {
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
          margin-bottom: 60px;
          text-shadow: 0 2px 4px rgba(0, 0, 0, 0.8);
          font-weight: bold;
        }

        .create-game-form {
          max-width: 400px;
          margin: 0 auto;
        }

        .input-container {
          position: relative;
          display: flex;
          align-items: center;
          background: rgba(255, 255, 255, 0.1);
          border: 2px solid rgba(255, 255, 255, 0.2);
          border-radius: 12px;
          padding: 0;
          transition: all 0.2s ease;
          backdrop-filter: blur(10px);
        }

        .input-container:focus-within {
          border-color: #4a90e2;
          box-shadow: 0 0 20px rgba(74, 144, 226, 0.3);
        }

        .player-name-input {
          flex: 1;
          background: transparent;
          border: none;
          padding: 20px 24px;
          color: white;
          font-size: 18px;
          outline: none;
          border-radius: 12px 0 0 12px;
        }

        .player-name-input::placeholder {
          color: rgba(255, 255, 255, 0.6);
        }

        .player-name-input:disabled {
          opacity: 0.6;
        }

        .submit-button {
          background: linear-gradient(135deg, #4a90e2 0%, #5ba0f2 100%);
          border: none;
          padding: 20px 24px;
          cursor: pointer;
          border-radius: 0 12px 12px 0;
          display: flex;
          align-items: center;
          justify-content: center;
          transition: all 0.2s ease;
          min-width: 80px;
        }

        .submit-button:hover:not(:disabled) {
          background: linear-gradient(135deg, #357abd 0%, #4a90e2 100%);
          transform: translateX(2px);
        }

        .submit-button:disabled {
          background: rgba(100, 100, 100, 0.5);
          cursor: not-allowed;
          transform: none;
        }

        .arrow-icon {
          width: 24px;
          height: 24px;
          filter: brightness(0) invert(1);
        }

        .submit-button:disabled .arrow-icon {
          opacity: 0.6;
        }

        .error-message {
          color: #ff6b6b;
          margin-top: 16px;
          padding: 12px;
          background: rgba(255, 107, 107, 0.1);
          border: 1px solid rgba(255, 107, 107, 0.3);
          border-radius: 8px;
          font-size: 14px;
        }

        .loading-message {
          color: #4a90e2;
          margin-top: 16px;
          padding: 12px;
          background: rgba(74, 144, 226, 0.1);
          border: 1px solid rgba(74, 144, 226, 0.3);
          border-radius: 8px;
          font-size: 14px;
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

        @media (max-width: 768px) {
          .container {
            padding: 20px 15px;
          }

          h1 {
            font-size: 36px;
            margin-bottom: 40px;
          }

          .player-name-input {
            padding: 16px 20px;
            font-size: 16px;
          }

          .submit-button {
            padding: 16px 20px;
            min-width: 70px;
          }

          .arrow-icon {
            width: 20px;
            height: 20px;
          }
        }

        @media (max-width: 480px) {
          h1 {
            font-size: 28px;
            margin-bottom: 30px;
          }

          .player-name-input {
            padding: 14px 18px;
            font-size: 14px;
          }

          .submit-button {
            padding: 14px 18px;
            min-width: 60px;
          }

          .arrow-icon {
            width: 18px;
            height: 18px;
          }
        }
      `}</style>
    </div>
  );
};

export default CreateGamePage;
