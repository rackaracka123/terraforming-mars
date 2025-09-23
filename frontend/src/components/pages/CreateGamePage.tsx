import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import { apiService } from "../../services/apiService";
import { globalWebSocketManager } from "../../services/globalWebSocketManager";
import { GameSettingsDto } from "../../types/generated/api-types.ts";
import styles from "./CreateGamePage.module.css";

const CreateGamePage: React.FC = () => {
  const navigate = useNavigate();
  const [playerName, setPlayerName] = useState("");
  const [developmentMode, setDevelopmentMode] = useState(true);
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
      const gameSettings: GameSettingsDto = {
        maxPlayers: 4, // Default max players
        developmentMode: developmentMode,
      };

      const game = await apiService.createGame(gameSettings);

      // Connect player to the game via WebSocket
      const playerConnectedResult = await globalWebSocketManager.playerConnect(
        playerName.trim(),
        game.id,
      );

      if (playerConnectedResult.game) {
        const gameData = {
          gameId: playerConnectedResult.game.id,
          playerId: playerConnectedResult.playerId,
          playerName: playerName.trim(),
          createdAt: new Date().toISOString(),
        };
        localStorage.setItem(
          "terraforming-mars-game",
          JSON.stringify(gameData),
        );

        // Navigate to the main game interface with the complete game state
        navigate("/game", {
          state: {
            game: playerConnectedResult.game,
            playerId: playerConnectedResult.playerId,
            playerName: playerName.trim(),
          },
        });
      } else {
        setError("Failed to connect to the game");
      }
    } catch (err) {
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
      void handleSubmit(e as React.FormEvent);
    }
  };

  return (
    <div className={styles.createGamePage}>
      <div className={styles.container}>
        <div className={styles.content}>
          <h1 className={styles.title}>Create a new game</h1>

          <form onSubmit={handleSubmit} className={styles.createGameForm}>
            <div className={styles.inputContainer}>
              <input
                type="text"
                value={playerName}
                onChange={handleInputChange}
                onKeyDown={handleKeyDown}
                placeholder="Enter your name"
                disabled={isLoading}
                className={styles.playerNameInput}
                autoFocus
                maxLength={50}
              />

              <button
                type="submit"
                disabled={isLoading || !playerName.trim()}
                className={styles.submitButton}
                title="Connect"
              >
                <img
                  src="/assets/misc/arrow.png"
                  alt="Connect"
                  className={styles.arrowIcon}
                />
              </button>
            </div>

            <div className={styles.developmentModeContainer}>
              <label className={styles.developmentModeLabel}>
                <input
                  type="checkbox"
                  checked={developmentMode}
                  onChange={(e) => setDevelopmentMode(e.target.checked)}
                  disabled={isLoading}
                  className={styles.developmentModeCheckbox}
                />
                <span className={styles.checkboxText}>
                  Development Mode
                  <div className={styles.tooltipContainer}>
                    <span className={styles.tooltipIcon}>ⓘ</span>
                    <div className={styles.tooltipText}>
                      Enable admin commands for debugging and testing. Allows
                      you to give cards to players, modify resources/production,
                      change game phases, and adjust global parameters through
                      the debug panel.
                    </div>
                  </div>
                </span>
              </label>
            </div>

            {error && <div className={styles.errorMessage}>{error}</div>}

            {isLoading && (
              <div className={styles.loadingMessage}>Creating game...</div>
            )}
          </form>
        </div>
      </div>
    </div>
  );
};

export default CreateGamePage;
