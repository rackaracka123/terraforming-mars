import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import { apiService } from "../../services/apiService";
import { webSocketService } from "../../services/webSocketService";

// UUIDv4 validation regex
const UUID_V4_REGEX = /^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;

const JoinGamePage: React.FC = () => {
  const navigate = useNavigate();
  const [gameId, setGameId] = useState("");
  const [playerName, setPlayerName] = useState("");
  const [isLoadingGameValidation, setIsLoadingGameValidation] = useState(false);
  const [isLoadingJoin, setIsLoadingJoin] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [gameValidated, setGameValidated] = useState(false);
  const [validatedGame, setValidatedGame] = useState<any>(null);

  const handleGameIdSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    // Validate UUID format
    if (!gameId.trim()) {
      setError("Please enter a game ID");
      return;
    }

    if (!UUID_V4_REGEX.test(gameId.trim())) {
      setError("Please enter a valid game ID (UUID format)");
      return;
    }

    setIsLoadingGameValidation(true);
    setError(null);

    try {
      // Verify the game exists and is joinable
      const game = await apiService.getGame(gameId.trim());

      // Check if game is full
      if (
        game.players &&
        game.players.length >= (game.settings?.maxPlayers || 4)
      ) {
        throw new Error("Game is full");
      }

      // Check if game is in a joinable state
      if (game.status !== "lobby" && game.status !== "waiting") {
        throw new Error("Game has already started");
      }

      // Game is valid and joinable
      setValidatedGame(game);
      setGameValidated(true);
    } catch (err) {
      if (err instanceof Error) {
        if (err.message.includes("404") || err.message.includes("not found")) {
          setError("Game not found. Please check the game ID.");
        } else if (err.message.includes("full")) {
          setError("This game is full. Please try another game.");
        } else if (err.message.includes("started")) {
          setError("This game has already started.");
        } else {
          setError(err.message || "Failed to find game");
        }
      } else {
        setError("Failed to find game");
      }
    } finally {
      setIsLoadingGameValidation(false);
    }
  };

  const handlePlayerNameSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    // Validation
    if (!playerName.trim()) {
      setError("Please enter your name");
      return;
    }

    if (playerName.trim().length < 2) {
      setError("Name must be at least 2 characters long");
      return;
    }

    setIsLoadingJoin(true);
    setError(null);

    try {
      // Connect to WebSocket if not already connected
      if (!webSocketService.connected) {
        await webSocketService.connect();
      }

      // Connect player to the game
      const playerConnectedResult = await webSocketService.playerConnect(
        playerName.trim(),
        validatedGame.id,
      );

      // Save game data to localStorage for reconnection
      const gameData = {
        gameId: validatedGame.id,
        playerId: playerConnectedResult.playerId,
        playerName: playerName.trim(),
        joinedAt: new Date().toISOString(),
      };
      localStorage.setItem("terraforming-mars-game", JSON.stringify(gameData));

      // Navigate to the main game interface
      navigate("/game", {
        state: {
          game: validatedGame,
          playerId: playerConnectedResult.playerId,
          playerName: playerName.trim(),
        },
      });
    } catch (err) {
      if (err instanceof Error) {
        if (err.message.includes("WebSocket")) {
          setError("Connection failed. Please try again.");
        } else {
          setError(err.message || "Failed to join game");
        }
      } else {
        setError("Failed to join game");
      }
    } finally {
      setIsLoadingJoin(false);
    }
  };

  const handleGameIdChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setGameId(e.target.value); // No transformation - keep original input
    if (error) setError(null);
  };

  const handlePlayerNameChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setPlayerName(e.target.value);
    if (error) setError(null);
  };

  const handleGameIdKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Enter") {
      handleGameIdSubmit(e as React.FormEvent);
    }
  };

  const handlePlayerNameKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Enter") {
      handlePlayerNameSubmit(e as React.FormEvent);
    }
  };

  const handleEditGameId = () => {
    setGameValidated(false);
    setValidatedGame(null);
    setPlayerName("");
    setError(null);
  };

  return (
    <div className="join-game-page">
      <div className="container">
        <div className="content">
          <h1>Join a game</h1>

          {!gameValidated ? (
            <form onSubmit={handleGameIdSubmit} className="join-game-form">
              <div className="input-container">
                <input
                  type="text"
                  value={gameId}
                  onChange={handleGameIdChange}
                  onKeyDown={handleGameIdKeyDown}
                  placeholder="Enter game ID"
                  disabled={isLoadingGameValidation}
                  className="player-name-input"
                  autoFocus
                />
                <button
                  type="submit"
                  disabled={isLoadingGameValidation || !gameId.trim()}
                  className="submit-button"
                  title="Find Game"
                >
                  <img
                    src="/assets/misc/arrow.png"
                    alt="Find Game"
                    className="arrow-icon"
                  />
                </button>
              </div>

              {error && <div className="error-message">{error}</div>}

              {isLoadingGameValidation && (
                <div className="loading-message">Finding game...</div>
              )}
            </form>
          ) : (
            <form onSubmit={handlePlayerNameSubmit} className="join-game-form">
              <div className="game-info">
                <p>
                  Game found! 
                  <button 
                    type="button" 
                    onClick={handleEditGameId}
                    className="edit-game-id-button"
                  >
                    (edit)
                  </button>
                </p>
              </div>

              <div className="input-container">
                <input
                  type="text"
                  value={playerName}
                  onChange={handlePlayerNameChange}
                  onKeyDown={handlePlayerNameKeyDown}
                  placeholder="Enter your name"
                  disabled={isLoadingJoin}
                  className="player-name-input"
                  autoFocus
                  maxLength={50}
                />
                <button
                  type="submit"
                  disabled={isLoadingJoin || !playerName.trim()}
                  className="submit-button"
                  title="Join Game"
                >
                  <img
                    src="/assets/misc/arrow.png"
                    alt="Join Game"
                    className="arrow-icon"
                  />
                </button>
              </div>

              {error && <div className="error-message">{error}</div>}

              {isLoadingJoin && (
                <div className="loading-message">Joining game...</div>
              )}
            </form>
          )}
        </div>
      </div>

      <style jsx>{`
        .join-game-page {
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

        .join-game-form {
          max-width: 400px;
          margin: 0 auto;
        }

        .game-info {
          margin-bottom: 30px;
        }

        .game-info p {
          color: rgba(255, 255, 255, 0.8);
          font-size: 16px;
          margin: 0;
        }

        .edit-game-id-button {
          background: transparent;
          border: none;
          color: #4a90e2;
          cursor: pointer;
          font-size: 14px;
          margin-left: 5px;
          text-decoration: underline;
        }

        .edit-game-id-button:hover {
          color: #5ba0f2;
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

export default JoinGamePage;
