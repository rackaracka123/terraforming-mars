import React, { useEffect, useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { apiService } from "../../services/apiService";
import { globalWebSocketManager } from "../../services/globalWebSocketManager";
import { skyboxCache } from "../../services/SkyboxCache.ts";
import styles from "./JoinGamePage.module.css";

// UUIDv4 validation regex
const UUID_V4_REGEX =
  /^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;

const JoinGamePage: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const [gameId, setGameId] = useState("");
  const [playerName, setPlayerName] = useState("");
  const [isLoadingGameValidation, setIsLoadingGameValidation] = useState(false);
  const [isLoadingJoin, setIsLoadingJoin] = useState(false);
  const [loadingStep, setLoadingStep] = useState<"game" | "environment" | null>(
    null,
  );
  const [error, setError] = useState<string | null>(null);
  const [gameValidated, setGameValidated] = useState(false);
  const [validatedGame, setValidatedGame] = useState<any>(null);
  const [skyboxReady, setSkyboxReady] = useState(false);
  const [isFadedIn, setIsFadedIn] = useState(false);

  // Check if skybox is already loaded on component mount
  useEffect(() => {
    if (skyboxCache.isReady()) {
      setSkyboxReady(true);
    }
    // Trigger fade in animation
    setTimeout(() => {
      setIsFadedIn(true);
    }, 10);
  }, []);

  // Handle URL parameter on mount
  useEffect(() => {
    const urlParams = new URLSearchParams(location.search);
    const codeParam = urlParams.get("code");

    if (codeParam && UUID_V4_REGEX.test(codeParam)) {
      setGameId(codeParam);
      // Auto-validate the game ID from URL
      void validateGameFromUrl(codeParam);
    }
  }, [location.search]);

  const validateGameFromUrl = async (gameIdFromUrl: string) => {
    setIsLoadingGameValidation(true);
    setError(null);

    try {
      const game = await apiService.getGame(gameIdFromUrl);
      if (!game) {
        throw new Error("Game not found");
      }

      // Check if game is full
      if (
        (game.currentPlayer ? 1 : 0) + (game.otherPlayers?.length || 0) >=
        (game.settings?.maxPlayers || 4)
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
          return;
        } else if (err.message.includes("full")) {
          setError("This game is full. Please try another game.");
          return;
        } else if (err.message.includes("started")) {
          setError("This game has already started.");
          return;
        } else {
          setError(err.message || "Failed to find game");
          return;
        }
      }
      setError("Failed to find game");
    } finally {
      setIsLoadingGameValidation(false);
    }
  };

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
      if (!game) {
        throw new Error("Game not found");
      }

      // Check if game is full
      if (
        (game.currentPlayer ? 1 : 0) + (game.otherPlayers?.length || 0) >=
        (game.settings?.maxPlayers || 4)
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
    setLoadingStep("game");

    try {
      // Step 1: Connect player to the game
      const playerConnectedResult = await globalWebSocketManager.playerConnect(
        playerName.trim(),
        validatedGame.id,
      );

      if (playerConnectedResult.game) {
        // Step 2: Load 3D environment if not already loaded
        if (!skyboxReady) {
          setLoadingStep("environment");
          await skyboxCache.preload();
        }

        const gameData = {
          gameId: validatedGame.id,
          playerId: playerConnectedResult.playerId,
          playerName: playerName.trim(),
          joinedAt: new Date().toISOString(),
        };
        localStorage.setItem(
          "terraforming-mars-game",
          JSON.stringify(gameData),
        );

        navigate("/game", {
          state: {
            game: playerConnectedResult.game,
            playerId: playerConnectedResult.playerId,
            playerName: playerName.trim(),
          },
        });
      } else {
        setError("Failed to join the game. Please try again.");
      }
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
      setLoadingStep(null);
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
      void handleGameIdSubmit(e as React.FormEvent);
    }
  };

  const handlePlayerNameKeyDown = (
    e: React.KeyboardEvent<HTMLInputElement>,
  ) => {
    if (e.key === "Enter") {
      void handlePlayerNameSubmit(e as React.FormEvent);
    }
  };

  const handleEditGameId = () => {
    setGameValidated(false);
    setValidatedGame(null);
    setPlayerName("");
    setError(null);
  };

  return (
    <div className={styles.joinGamePage} style={{
      opacity: isFadedIn ? 1 : 0,
      transition: "opacity 0.3s ease-in",
    }}>
        <div className={styles.container}>
          <div className={styles.content}>
            <h1>Join a game</h1>

            {!gameValidated ? (
              <form onSubmit={handleGameIdSubmit} className={styles.joinGameForm}>
                <div className={styles.inputContainer}>
                  <input
                    type="text"
                    value={gameId}
                    onChange={handleGameIdChange}
                    onKeyDown={handleGameIdKeyDown}
                    placeholder="Enter game ID"
                    disabled={isLoadingGameValidation}
                    className={styles.playerNameInput}
                    autoFocus
                  />
                  <button
                    type="submit"
                    disabled={isLoadingGameValidation || !gameId.trim()}
                    className={styles.submitButton}
                    title="Find Game"
                  >
                    <img
                      src="/assets/misc/arrow.png"
                      alt="Find Game"
                      className={styles.arrowIcon}
                    />
                  </button>
                </div>

                {error && <div className={styles.errorMessage}>{error}</div>}

                {isLoadingGameValidation && (
                  <div className={styles.loadingMessage}>Finding game...</div>
                )}
              </form>
            ) : (
              <form
                onSubmit={handlePlayerNameSubmit}
                className={styles.joinGameForm}
              >
                <div className={styles.gameInfo}>
                  <p>
                    Game found!
                    <button
                      type="button"
                      onClick={handleEditGameId}
                      className={styles.editGameIdButton}
                    >
                      (edit)
                    </button>
                  </p>
                </div>

                <div className={styles.inputContainer}>
                  <input
                    type="text"
                    value={playerName}
                    onChange={handlePlayerNameChange}
                    onKeyDown={handlePlayerNameKeyDown}
                    placeholder="Enter your name"
                    disabled={isLoadingJoin}
                    className={styles.playerNameInput}
                    autoFocus
                    maxLength={50}
                  />
                  <button
                    type="submit"
                    disabled={isLoadingJoin || !playerName.trim()}
                    className={styles.submitButton}
                    title="Join Game"
                  >
                    <img
                      src="/assets/misc/arrow.png"
                      alt="Join Game"
                      className={styles.arrowIcon}
                    />
                  </button>
                </div>

                {error && <div className={styles.errorMessage}>{error}</div>}

                {isLoadingJoin && (
                  <div className={styles.loadingMessage}>
                    {loadingStep === "game" && "Joining game..."}
                    {loadingStep === "environment" && "Loading 3D environment..."}
                  </div>
                )}
              </form>
            )}
          </div>
        </div>
      </div>
  );
};

export default JoinGamePage;
